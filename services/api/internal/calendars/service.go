// Package calendars manages external calendar accounts and collections.
package calendars

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/calendars/caldav"
	"github.com/kurtschwarz/homie/services/api/internal/secrets"
)

const (
	ProviderCalDAV      = "caldav"
	MinimumRefresh      = 5 * time.Minute
	DefaultRefresh      = 15 * time.Minute
	DefaultWindowPast   = 24 * time.Hour
	DefaultWindowFuture = 21 * 24 * time.Hour
)

var (
	ErrNotFound       = errors.New("calendar not found")
	ErrInvalidAccount = errors.New("invalid calendar account")
	ErrReadOnly       = errors.New("calendar is read-only")
)

// CalDAVConnection is how to reach and authenticate to a CalDAV server.
// Password is write-only: reads never populate it.
type CalDAVConnection struct {
	URL, Username, Password, Timezone string
}

// Account is one login at a calendar provider. The provider-specific part
// lives in the field named for the provider, which is set when Provider matches.
type Account struct {
	ID                                        uuid.UUID
	Name, Provider                            string
	CalDAV                                    *CalDAVConnection
	RefreshInterval, WindowPast, WindowFuture time.Duration
	LastAttemptedAt                           *time.Time
}
type storedAccount struct {
	Account
	Credentials []byte
}

// Calendar is one provider collection belonging to an Account.
type Calendar struct {
	ID, AccountID     uuid.UUID
	ProviderRef, Name string
	ReadOnly, Visible bool
	Order             int
	LastSyncedAt      *time.Time
	LastError         string
}

type Occurrence struct {
	UID, RecurrenceID, Summary, Location string
	Start, End                           time.Time
	AllDay                               bool
	Categories                           []string
}
type Client interface {
	Report(context.Context, time.Time, time.Time) ([]Occurrence, error)
	Put(context.Context, Occurrence) error
}
type Source struct {
	Account  Account
	Calendar Calendar
	Client   Client
}

type Repository interface {
	ListAccounts(context.Context) ([]storedAccount, error)
	GetAccount(context.Context, uuid.UUID) (storedAccount, error)
	UpsertAccount(context.Context, storedAccount) error
	DeleteAccount(context.Context, uuid.UUID) error
	ListCalendars(context.Context, *uuid.UUID) ([]Calendar, error)
	GetCalendar(context.Context, uuid.UUID) (Calendar, error)
	UpsertCalendar(context.Context, Calendar) error
	UpsertDiscoveredCalendar(context.Context, Calendar) error
	SetCalendarVisible(context.Context, uuid.UUID, bool) error
	RecordSync(context.Context, uuid.UUID, *time.Time, string) error
	RecordAttempt(context.Context, uuid.UUID, time.Time) error
}

type Service struct {
	repo   Repository
	box    *secrets.Box
	logger *slog.Logger
}

func New(repo Repository, box *secrets.Box, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{repo: repo, box: box, logger: logger}
}

func (s *Service) ListAccounts(ctx context.Context) ([]Account, error) {
	rows, err := s.repo.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Account, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.Account)
	}
	return result, nil
}
func (s *Service) ListCalendars(ctx context.Context, accountID *uuid.UUID) ([]Calendar, error) {
	return s.repo.ListCalendars(ctx, accountID)
}

// UpdateCalendarVisibility shows or hides a calendar's events on the display.
// Hidden calendars keep syncing, so showing one takes effect immediately.
func (s *Service) UpdateCalendarVisibility(ctx context.Context, id uuid.UUID, visible bool) (Calendar, error) {
	if err := s.repo.SetCalendarVisible(ctx, id, visible); err != nil {
		return Calendar{}, err
	}
	return s.repo.GetCalendar(ctx, id)
}

func (s *Service) CreateAccount(ctx context.Context, account Account) (Account, error) {
	account.ID = uuid.Must(uuid.NewV7())
	applyAccountDefaults(&account)
	if err := validateAccount(account); err != nil {
		return Account{}, err
	}
	if account.CalDAV.Password == "" {
		return Account{}, fmt.Errorf("%w: password is required", ErrInvalidAccount)
	}
	encrypted, err := s.sealPassword(account.CalDAV.Password)
	if err != nil {
		return Account{}, err
	}
	connection := *account.CalDAV
	connection.Password = ""
	account.CalDAV = &connection
	if err := s.repo.UpsertAccount(ctx, storedAccount{Account: account, Credentials: encrypted}); err != nil {
		return Account{}, err
	}
	return account, nil
}

// CalDAVPatch carries the CalDAV fields an update may change; nil leaves a field as is.
type CalDAVPatch struct {
	URL, Username, Password, Timezone *string
}

// AccountPatch carries the fields an update may change; nil leaves a field as is.
type AccountPatch struct {
	ID                                        uuid.UUID
	Name                                      *string
	CalDAV                                    *CalDAVPatch
	RefreshInterval, WindowPast, WindowFuture *time.Duration
}

// UpdateAccount applies a patch. Changing the URL or username requires a new
// password, so stored credentials are never sent to an endpoint the caller
// chose without knowing them.
func (s *Service) UpdateAccount(ctx context.Context, patch AccountPatch) (Account, error) {
	stored, err := s.repo.GetAccount(ctx, patch.ID)
	if err != nil {
		return Account{}, err
	}
	account := stored.Account
	encrypted := stored.Credentials
	if account.CalDAV != nil {
		connection := *account.CalDAV
		account.CalDAV = &connection
	}
	if patch.Name != nil {
		account.Name = *patch.Name
	}
	if patch.RefreshInterval != nil {
		account.RefreshInterval = *patch.RefreshInterval
	}
	if patch.WindowPast != nil {
		account.WindowPast = *patch.WindowPast
	}
	if patch.WindowFuture != nil {
		account.WindowFuture = *patch.WindowFuture
	}
	if patch.CalDAV != nil {
		if account.CalDAV == nil {
			return Account{}, fmt.Errorf("%w: account is not a CalDAV account", ErrInvalidAccount)
		}
		changed := patch.CalDAV
		endpointChanged := false
		if changed.Timezone != nil {
			account.CalDAV.Timezone = *changed.Timezone
		}
		if changed.URL != nil && *changed.URL != account.CalDAV.URL {
			account.CalDAV.URL, endpointChanged = *changed.URL, true
		}
		if changed.Username != nil && *changed.Username != account.CalDAV.Username {
			account.CalDAV.Username, endpointChanged = *changed.Username, true
		}
		if changed.Password != nil && *changed.Password != "" {
			if encrypted, err = s.sealPassword(*changed.Password); err != nil {
				return Account{}, err
			}
		} else if endpointChanged {
			return Account{}, fmt.Errorf("%w: password is required when connection details change", ErrInvalidAccount)
		}
	}
	if err := validateAccount(account); err != nil {
		return Account{}, err
	}
	if err := s.repo.UpsertAccount(ctx, storedAccount{Account: account, Credentials: encrypted}); err != nil {
		return Account{}, err
	}
	return account, nil
}

func (s *Service) sealPassword(password string) ([]byte, error) {
	if s.box == nil {
		return nil, errors.New("calendar credential encryption is not configured")
	}
	plain, err := json.Marshal(struct {
		Password string `json:"password"`
	}{password})
	if err != nil {
		return nil, err
	}
	return s.box.Seal(plain)
}
func (s *Service) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteAccount(ctx, id)
}

// TestStoredAccount checks a saved account with its stored credentials.
func (s *Service) TestStoredAccount(ctx context.Context, id uuid.UUID) error {
	stored, err := s.repo.GetAccount(ctx, id)
	if err != nil {
		return err
	}
	connection, err := s.openConnection(stored)
	if err != nil {
		return err
	}
	return findCalendars(ctx, connection)
}

// TestDraftAccount checks a connection that has not been saved. It never
// uses stored credentials, so the caller must supply the password.
func (s *Service) TestDraftAccount(ctx context.Context, connection CalDAVConnection) error {
	draft := Account{Name: "connection test", Provider: ProviderCalDAV, CalDAV: &connection}
	applyAccountDefaults(&draft)
	if err := validateAccount(draft); err != nil {
		return err
	}
	if connection.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidAccount)
	}
	return findCalendars(ctx, connection)
}

func findCalendars(ctx context.Context, connection CalDAVConnection) error {
	client, err := newCalDAVClient(connection, "")
	if err != nil {
		return err
	}
	_, err = client.FindCalendars(ctx)
	return err
}

// Discover upserts every VEVENT collection and leaves new calendars disabled.
func (s *Service) Discover(ctx context.Context, accountID uuid.UUID) ([]Calendar, error) {
	stored, err := s.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	connection, err := s.openConnection(stored)
	if err != nil {
		return nil, err
	}
	client, err := newCalDAVClient(connection, "")
	if err != nil {
		return nil, err
	}
	found, err := client.FindCalendars(ctx)
	if err != nil {
		return nil, err
	}
	for index, item := range found {
		calendar := Calendar{ID: uuid.Must(uuid.NewV7()), AccountID: accountID, ProviderRef: item.Path, Name: item.Name, Order: index}
		if err := s.repo.UpsertDiscoveredCalendar(ctx, calendar); err != nil {
			return nil, err
		}
	}
	return s.repo.ListCalendars(ctx, &accountID)
}

func (s *Service) Sources(ctx context.Context) ([]Source, error) {
	calendars, err := s.repo.ListCalendars(ctx, nil)
	if err != nil {
		return nil, err
	}
	accounts := map[uuid.UUID]storedAccount{}
	var result []Source
	for _, calendar := range calendars {
		account, ok := accounts[calendar.AccountID]
		if !ok {
			account, err = s.repo.GetAccount(ctx, calendar.AccountID)
			if err != nil {
				s.recordBroken(ctx, calendar.ID, err.Error())
				continue
			}
			accounts[calendar.AccountID] = account
		}
		source, err := s.source(ctx, account, calendar)
		if err == nil {
			result = append(result, source)
		}
	}
	return result, nil
}
func (s *Service) Source(ctx context.Context, id uuid.UUID) (Source, error) {
	calendar, err := s.repo.GetCalendar(ctx, id)
	if err != nil {
		return Source{}, err
	}
	account, err := s.repo.GetAccount(ctx, calendar.AccountID)
	if err != nil {
		return Source{}, err
	}
	return s.source(ctx, account, calendar)
}
func (s *Service) source(ctx context.Context, stored storedAccount, calendar Calendar) (Source, error) {
	connection, err := s.openConnection(stored)
	if err != nil {
		s.recordBroken(ctx, calendar.ID, "stored credentials cannot be decrypted")
		return Source{}, err
	}
	client, err := newClient(connection, calendar)
	if err != nil {
		s.recordBroken(ctx, calendar.ID, err.Error())
		return Source{}, err
	}
	return Source{Account: stored.Account, Calendar: calendar, Client: client}, nil
}

// openConnection returns a copy of the stored connection with its password
// decrypted. The stored Account is left without a password.
func (s *Service) openConnection(stored storedAccount) (CalDAVConnection, error) {
	if stored.CalDAV == nil {
		return CalDAVConnection{}, fmt.Errorf("%w: unsupported provider %q", ErrInvalidAccount, stored.Provider)
	}
	password, err := s.decryptPassword(stored)
	if err != nil {
		return CalDAVConnection{}, err
	}
	connection := *stored.CalDAV
	connection.Password = password
	return connection, nil
}
func (s *Service) decryptPassword(stored storedAccount) (string, error) {
	if s.box == nil {
		return "", errors.New("credential encryption is not configured")
	}
	plain, err := s.box.Open(stored.Credentials)
	if err != nil {
		return "", err
	}
	var credentials struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(plain, &credentials); err != nil {
		return "", err
	}
	return credentials.Password, nil
}

func (s *Service) RecordSuccess(ctx context.Context, id uuid.UUID, at time.Time) error {
	return s.repo.RecordSync(ctx, id, &at, "")
}
func (s *Service) RecordFailure(ctx context.Context, id uuid.UUID, message string) error {
	return s.repo.RecordSync(ctx, id, nil, message)
}
func (s *Service) RecordAttempt(ctx context.Context, id uuid.UUID, at time.Time) error {
	return s.repo.RecordAttempt(ctx, id, at)
}
func (s *Service) recordBroken(ctx context.Context, id uuid.UUID, message string) {
	s.logger.WarnContext(ctx, "calendar connection unavailable", "calendar_id", id, "error", message)
	_ = s.RecordFailure(ctx, id, message)
}

func applyAccountDefaults(account *Account) {
	if account.RefreshInterval == 0 {
		account.RefreshInterval = DefaultRefresh
	}
	if account.WindowPast == 0 {
		account.WindowPast = DefaultWindowPast
	}
	if account.WindowFuture == 0 {
		account.WindowFuture = DefaultWindowFuture
	}
}
func validateAccount(account Account) error {
	if account.Name == "" || account.Provider != ProviderCalDAV || account.CalDAV == nil || account.RefreshInterval < MinimumRefresh || account.WindowPast < 0 || account.WindowFuture <= 0 {
		return ErrInvalidAccount
	}
	if account.CalDAV.URL == "" || account.CalDAV.Username == "" {
		return ErrInvalidAccount
	}
	if account.CalDAV.Timezone != "" {
		if _, err := time.LoadLocation(account.CalDAV.Timezone); err != nil {
			return ErrInvalidAccount
		}
	}
	return nil
}

type calDAVAdapter struct {
	client   *caldav.Client
	readOnly bool
}

func newCalDAVClient(connection CalDAVConnection, collectionPath string) (*caldav.Client, error) {
	location := time.UTC
	if connection.Timezone != "" {
		var err error
		location, err = time.LoadLocation(connection.Timezone)
		if err != nil {
			return nil, fmt.Errorf("load calendar timezone: %w", err)
		}
	}
	return caldav.New(caldav.Config{URL: connection.URL, User: connection.Username, Password: connection.Password, CollectionPath: collectionPath, Location: location})
}
func newClient(connection CalDAVConnection, calendar Calendar) (Client, error) {
	client, err := newCalDAVClient(connection, calendar.ProviderRef)
	if err != nil {
		return nil, err
	}
	return &calDAVAdapter{client: client, readOnly: calendar.ReadOnly}, nil
}
func (a *calDAVAdapter) Report(ctx context.Context, from, to time.Time) ([]Occurrence, error) {
	rows, err := a.client.Report(ctx, from, to)
	if err != nil {
		return nil, err
	}
	result := make([]Occurrence, 0, len(rows))
	for _, row := range rows {
		result = append(result, Occurrence{UID: row.UID, RecurrenceID: row.RecurrenceID, Summary: row.Summary, Location: row.Location, Start: row.Start, End: row.End, AllDay: row.AllDay, Categories: row.Categories})
	}
	return result, nil
}
func (a *calDAVAdapter) Put(ctx context.Context, occurrence Occurrence) error {
	if a.readOnly {
		return ErrReadOnly
	}
	return a.client.Put(ctx, caldav.Occurrence{UID: occurrence.UID, RecurrenceID: occurrence.RecurrenceID, Summary: occurrence.Summary, Location: occurrence.Location, Start: occurrence.Start, End: occurrence.End, AllDay: occurrence.AllDay, Categories: occurrence.Categories})
}

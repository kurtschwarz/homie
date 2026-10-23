package calendars

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/secrets"
)

type memoryRepository struct {
	accounts  map[uuid.UUID]storedAccount
	calendars map[uuid.UUID]Calendar
}

func (r *memoryRepository) ListAccounts(context.Context) ([]storedAccount, error) {
	result := make([]storedAccount, 0, len(r.accounts))
	for _, row := range r.accounts {
		result = append(result, row)
	}
	return result, nil
}
func (r *memoryRepository) GetAccount(_ context.Context, id uuid.UUID) (storedAccount, error) {
	row, ok := r.accounts[id]
	if !ok {
		return storedAccount{}, ErrNotFound
	}
	return row, nil
}
func (r *memoryRepository) UpsertAccount(_ context.Context, row storedAccount) error {
	r.accounts[row.ID] = row
	return nil
}
func (r *memoryRepository) DeleteAccount(_ context.Context, id uuid.UUID) error {
	delete(r.accounts, id)
	return nil
}
func (r *memoryRepository) ListCalendars(_ context.Context, accountID *uuid.UUID) ([]Calendar, error) {
	var result []Calendar
	for _, row := range r.calendars {
		if accountID == nil || row.AccountID == *accountID {
			result = append(result, row)
		}
	}
	return result, nil
}
func (r *memoryRepository) GetCalendar(_ context.Context, id uuid.UUID) (Calendar, error) {
	row, ok := r.calendars[id]
	if !ok {
		return Calendar{}, ErrNotFound
	}
	return row, nil
}
func (r *memoryRepository) UpsertCalendar(_ context.Context, row Calendar) error {
	r.calendars[row.ID] = row
	return nil
}
func (r *memoryRepository) UpsertDiscoveredCalendar(ctx context.Context, row Calendar) error {
	return r.UpsertCalendar(ctx, row)
}
func (r *memoryRepository) SetCalendarVisible(_ context.Context, id uuid.UUID, visible bool) error {
	row, ok := r.calendars[id]
	if !ok {
		return ErrNotFound
	}
	row.Visible = visible
	r.calendars[id] = row
	return nil
}
func (r *memoryRepository) RecordSync(context.Context, uuid.UUID, *time.Time, string) error {
	return nil
}
func (r *memoryRepository) RecordAttempt(context.Context, uuid.UUID, time.Time) error { return nil }

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{accounts: map[uuid.UUID]storedAccount{}, calendars: map[uuid.UUID]Calendar{}}
}

func newTestService(t *testing.T) (*memoryRepository, *secrets.Box, *Service) {
	t.Helper()
	repo := newMemoryRepository()
	box, err := secrets.New(bytes.Repeat([]byte{3}, 32))
	if err != nil {
		t.Fatal(err)
	}
	return repo, box, New(repo, box, nil)
}

func caldavAccount(url, password string) Account {
	return Account{Name: "Fastmail", Provider: ProviderCalDAV, CalDAV: &CalDAVConnection{URL: url, Username: "me", Password: password}}
}

func TestCreateAccountEncryptsPasswordAndUpdatePreservesIt(t *testing.T) {
	repo, box, service := newTestService(t)
	saved, err := service.CreateAccount(context.Background(), caldavAccount("https://example.test/calendar/", "app-password"))
	if err != nil {
		t.Fatal(err)
	}
	if saved.CalDAV.Password != "" {
		t.Fatal("password leaked through create")
	}
	row := repo.accounts[saved.ID]
	if bytes.Contains(row.Credentials, []byte("app-password")) {
		t.Fatal("password was stored in plaintext")
	}
	plain, err := box.Open(row.Credentials)
	if err != nil {
		t.Fatal(err)
	}
	var credential struct {
		Password string `json:"password"`
	}
	if err := json.Unmarshal(plain, &credential); err != nil {
		t.Fatal(err)
	}
	if credential.Password != "app-password" {
		t.Fatalf("got password %q", credential.Password)
	}
	renamed := "Renamed"
	if _, err := service.UpdateAccount(context.Background(), AccountPatch{ID: saved.ID, Name: &renamed}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(repo.accounts[saved.ID].Credentials, row.Credentials) {
		t.Fatal("password changed when update omitted it")
	}
	listed, err := service.ListAccounts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if listed[0].Name != "Renamed" || listed[0].CalDAV.Password != "" {
		t.Fatalf("unexpected listed account: %+v", listed[0])
	}
}

func TestCreateAccountRejectsTooFrequentRefresh(t *testing.T) {
	_, _, service := newTestService(t)
	account := caldavAccount("https://example.test/", "secret")
	account.RefreshInterval, account.WindowFuture = time.Minute, time.Hour
	if _, err := service.CreateAccount(context.Background(), account); !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateCalendarVisibility(t *testing.T) {
	repo, _, service := newTestService(t)
	ctx := context.Background()
	id := uuid.New()
	repo.calendars[id] = Calendar{ID: id, AccountID: uuid.New(), ProviderRef: "/home/family/", Name: "Family"}
	shown, err := service.UpdateCalendarVisibility(ctx, id, true)
	if err != nil || !shown.Visible {
		t.Fatalf("show: got=%+v err=%v", shown, err)
	}
	hidden, err := service.UpdateCalendarVisibility(ctx, id, false)
	if err != nil || hidden.Visible {
		t.Fatalf("hide: got=%+v err=%v", hidden, err)
	}
	if _, err := service.UpdateCalendarVisibility(ctx, uuid.New(), true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown calendar: got %v", err)
	}
}

func TestCreateAccountRequiresPassword(t *testing.T) {
	_, _, service := newTestService(t)
	if _, err := service.CreateAccount(context.Background(), caldavAccount("https://example.test/", "")); !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("got %v", err)
	}
}

func TestTestDraftAccountRequiresItsOwnPassword(t *testing.T) {
	_, _, service := newTestService(t)
	err := service.TestDraftAccount(context.Background(), CalDAVConnection{URL: "https://attacker.example/", Username: "me"})
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateAccountRequiresPasswordWhenEndpointChanges(t *testing.T) {
	repo, _, service := newTestService(t)
	saved, err := service.CreateAccount(context.Background(), caldavAccount("https://calendar.example/home/", "secret"))
	if err != nil {
		t.Fatal(err)
	}
	before := repo.accounts[saved.ID].Credentials
	moved := "https://attacker.example/"
	if _, err := service.UpdateAccount(context.Background(), AccountPatch{ID: saved.ID, CalDAV: &CalDAVPatch{URL: &moved}}); !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("got %v", err)
	}
	if repo.accounts[saved.ID].CalDAV.URL != "https://calendar.example/home/" {
		t.Fatal("endpoint changed without a new password")
	}
	fresh := "new-secret"
	if _, err := service.UpdateAccount(context.Background(), AccountPatch{ID: saved.ID, CalDAV: &CalDAVPatch{URL: &moved, Password: &fresh}}); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(repo.accounts[saved.ID].Credentials, before) {
		t.Fatal("password was not replaced")
	}
}

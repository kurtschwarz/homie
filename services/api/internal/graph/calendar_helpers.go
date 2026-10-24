package graph

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kurtschwarz/homie/services/api/internal/calendars"
	"github.com/kurtschwarz/homie/services/api/internal/graph/model"
)

// These live here rather than beside the resolvers because gqlgen comments
// out anything it did not generate when it rewrites the resolver file.
var (
	errCalendarsNotConfigured = errors.New("calendars are not configured")
	errConnectionProvider     = errors.New("connection must name a supported provider")
)

func caldavFromInput(input model.CalDAVConnectionInput) *calendars.CalDAVConnection {
	connection := &calendars.CalDAVConnection{URL: input.URL, Username: input.Username, Password: input.Password}
	if input.Timezone != nil {
		connection.Timezone = *input.Timezone
	}
	return connection
}

func accountFromCreateInput(input model.CreateCalendarAccountInput) (calendars.Account, error) {
	if input.Connection == nil || input.Connection.Caldav == nil {
		return calendars.Account{}, errConnectionProvider
	}
	return calendars.Account{
		Name: input.Name, Provider: calendars.ProviderCalDAV,
		CalDAV:          caldavFromInput(*input.Connection.Caldav),
		RefreshInterval: seconds(input.RefreshIntervalSeconds),
		WindowPast:      seconds(input.WindowPastSeconds),
		WindowFuture:    seconds(input.WindowFutureSeconds),
	}, nil
}

func patchFromUpdateInput(input model.UpdateCalendarAccountInput) (calendars.AccountPatch, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return calendars.AccountPatch{}, err
	}
	patch := calendars.AccountPatch{
		ID: id, Name: input.Name,
		RefreshInterval: secondsPtr(input.RefreshIntervalSeconds),
		WindowPast:      secondsPtr(input.WindowPastSeconds),
		WindowFuture:    secondsPtr(input.WindowFutureSeconds),
	}
	if input.Connection != nil {
		if input.Connection.Caldav == nil {
			return calendars.AccountPatch{}, errConnectionProvider
		}
		changed := input.Connection.Caldav
		patch.CalDAV = &calendars.CalDAVPatch{URL: changed.URL, Username: changed.Username, Password: changed.Password, Timezone: changed.Timezone}
	}
	return patch, nil
}

func seconds(value int) time.Duration { return time.Duration(value) * time.Second }

func secondsPtr(value *int) *time.Duration {
	if value == nil {
		return nil
	}
	duration := seconds(*value)
	return &duration
}

func accountToModel(account calendars.Account) model.CalendarAccount {
	result := model.CalendarAccount{
		ID: account.ID.String(), Name: account.Name, Provider: model.CalendarProviderCaldav,
		RefreshIntervalSeconds: int(account.RefreshInterval / time.Second),
		WindowPastSeconds:      int(account.WindowPast / time.Second),
		WindowFutureSeconds:    int(account.WindowFuture / time.Second),
	}
	if account.CalDAV != nil {
		result.Connection = model.CalDAVConnection{URL: account.CalDAV.URL, Username: account.CalDAV.Username, Timezone: account.CalDAV.Timezone}
	}
	return result
}

func calendarToModel(calendar calendars.Calendar) model.Calendar {
	result := model.Calendar{
		ID: calendar.ID.String(), AccountID: calendar.AccountID.String(), ProviderRef: calendar.ProviderRef,
		Name: calendar.Name, ReadOnly: calendar.ReadOnly, Visible: calendar.Visible, LastError: calendar.LastError,
	}
	if calendar.LastSyncedAt != nil {
		value := int(calendar.LastSyncedAt.Unix())
		result.LastSyncedAt = &value
	}
	return result
}

func calendarsToModel(rows []calendars.Calendar) []model.Calendar {
	result := make([]model.Calendar, 0, len(rows))
	for _, row := range rows {
		result = append(result, calendarToModel(row))
	}
	return result
}

func (r *Resolver) calendarsWithCounts(ctx context.Context, rows []calendars.Calendar) ([]model.Calendar, error) {
	result := calendarsToModel(rows)
	if r.Events == nil {
		return result, nil
	}
	counts, err := r.Events.UpcomingCalendarEventCounts(ctx)
	if err != nil {
		return nil, err
	}
	for i := range result {
		result[i].EventCount = counts[rows[i].ID]
	}
	return result, nil
}

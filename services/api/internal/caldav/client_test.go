package caldav

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const multistatusFixture = `<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
  <d:response>
    <d:href>/dav/calendars/user/kurt@schw.ar/19abcc4c/mortgage.ics</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"1"</d:getetag>
        <c:calendar-data>BEGIN:VCALENDAR
BEGIN:VEVENT
UID:mortgage-1
DTSTART;VALUE=DATE:20261019
SUMMARY:Mortgage Payment - $2\,248.08
END:VEVENT
END:VCALENDAR
</c:calendar-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dav/calendars/user/kurt@schw.ar/19abcc4c/school.ics</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"2"</d:getetag>
        <c:calendar-data>BEGIN:VCALENDAR
BEGIN:VEVENT
UID:school-pickup
RECURRENCE-ID;TZID=America/Toronto:20261015T150000
DTSTART;TZID=America/Toronto:20261015T090000
DTEND;TZID=America/Toronto:20261015T100000
SUMMARY:School 
 pickup
CATEGORIES:Elina, Yellow
END:VEVENT
END:VCALENDAR
</c:calendar-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>/dav/calendars/user/kurt@schw.ar/19abcc4c/duration.ics</d:href>
    <d:propstat>
      <d:prop>
        <c:calendar-data>BEGIN:VCALENDAR
BEGIN:VEVENT
UID:swim-lessons
DTSTART:20261016T170000Z
DURATION:PT1H30M
SUMMARY:Swim lessons
END:VEVENT
END:VCALENDAR
</c:calendar-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`

func TestReportParsesMultistatusIntoOccurrences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "REPORT" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(multistatusFixture))
	}))
	defer server.Close()

	client, err := New(Config{URL: server.URL, User: "kurt@schw.ar", Password: "app-password", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	occurrences, err := client.Report(context.Background(), time.Now(), time.Now().AddDate(0, 0, 21))
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 3 {
		t.Fatalf("expected 3 occurrences, got %d: %+v", len(occurrences), occurrences)
	}
}

// The expand element is what makes every other assumption in this package
// hold: without it the server returns RRULE masters and nothing here would
// know the difference, because a master parses into a perfectly valid-looking
// Occurrence dated at the series' first instance.
func TestReportRequestsServerSideExpansion(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(multistatusFixture))
	}))
	defer server.Close()

	client, err := New(Config{URL: server.URL, User: "kurt@schw.ar", Password: "app-password", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Report(context.Background(), time.Now(), time.Now().AddDate(0, 0, 21)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "expand") {
		t.Fatalf("REPORT body does not ask the server to expand recurrences: %s", body)
	}
}

func TestReportUnfoldsFoldedSummary(t *testing.T) {
	occurrences := mustParseFixture(t)
	var school *Occurrence
	for i := range occurrences {
		if occurrences[i].UID == "school-pickup" {
			school = &occurrences[i]
		}
	}
	if school == nil || school.Summary != "School pickup" {
		t.Fatalf("expected unfolded summary 'School pickup', got %+v", school)
	}
}

func TestReportUnescapesSummary(t *testing.T) {
	occurrences := mustParseFixture(t)
	var mortgage *Occurrence
	for i := range occurrences {
		if occurrences[i].UID == "mortgage-1" {
			mortgage = &occurrences[i]
		}
	}
	if mortgage == nil {
		t.Fatal("mortgage occurrence not found")
	}
	if !strings.Contains(mortgage.Summary, "$2,248.08") {
		t.Fatalf("expected an unescaped comma in the amount, got %q", mortgage.Summary)
	}
}

func TestReportSplitsAndNormalisesCategories(t *testing.T) {
	occurrences := mustParseFixture(t)
	var school *Occurrence
	for i := range occurrences {
		if occurrences[i].UID == "school-pickup" {
			school = &occurrences[i]
		}
	}
	if school == nil || len(school.Categories) != 2 || school.Categories[0] != "elina" || school.Categories[1] != "yellow" {
		t.Fatalf("expected normalised categories [elina yellow], got %+v", school)
	}
}

func TestReportReadsAllDayEventToLocalMidnight(t *testing.T) {
	occurrences := mustParseFixture(t)
	var mortgage *Occurrence
	for i := range occurrences {
		if occurrences[i].UID == "mortgage-1" {
			mortgage = &occurrences[i]
		}
	}
	if mortgage == nil || !mortgage.AllDay {
		t.Fatalf("expected an all-day event, got %+v", mortgage)
	}
	if mortgage.Start.Hour() != 0 || mortgage.Start.Minute() != 0 {
		t.Fatalf("expected local midnight, got %s", mortgage.Start)
	}
}

func TestReportReadsTZIDEventToCorrectInstant(t *testing.T) {
	occurrences := mustParseFixture(t)
	var school *Occurrence
	for i := range occurrences {
		if occurrences[i].UID == "school-pickup" {
			school = &occurrences[i]
		}
	}
	if school == nil {
		t.Fatal("school occurrence not found")
	}
	toronto, err := time.LoadLocation("America/Toronto")
	if err != nil {
		t.Skipf("tzdata unavailable: %v", err)
	}
	want := time.Date(2026, 10, 15, 9, 0, 0, 0, toronto)
	if !school.Start.Equal(want) {
		t.Fatalf("expected %s, got %s", want, school.Start)
	}
	if school.RecurrenceID == "" {
		t.Fatal("expected an expanded instance to keep its RECURRENCE-ID as identity")
	}
}

func TestReportDerivesEndFromDuration(t *testing.T) {
	occurrences := mustParseFixture(t)
	var swim *Occurrence
	for i := range occurrences {
		if occurrences[i].UID == "swim-lessons" {
			swim = &occurrences[i]
		}
	}
	if swim == nil {
		t.Fatal("swim occurrence not found")
	}
	if got, want := swim.End.Sub(swim.Start), 90*time.Minute; got != want {
		t.Fatalf("expected a 90 minute duration-derived end, got %s", got)
	}
}

func TestNewAppendsATrailingSlashToTheCollectionURL(t *testing.T) {
	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(`<d:multistatus xmlns:d="DAV:"></d:multistatus>`))
	}))
	defer server.Close()
	client, err := New(Config{URL: server.URL + "/dav/calendars/user/kurt", User: "kurt", Password: "pw", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Report(context.Background(), time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if requestedPath != "/dav/calendars/user/kurt/" {
		t.Fatalf("expected the request path to carry the trailing slash, got %q", requestedPath)
	}
}

func TestReportReturnsDistinguishableUnauthorizedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()
	client, err := New(Config{URL: server.URL, User: "kurt@schw.ar", Password: "revoked", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Report(context.Background(), time.Now(), time.Now().AddDate(0, 0, 21))
	if err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func mustParseFixture(t *testing.T) []Occurrence {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(multistatusFixture))
	}))
	defer server.Close()
	client, err := New(Config{URL: server.URL, User: "kurt@schw.ar", Password: "app-password", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	occurrences, err := client.Report(context.Background(), time.Now(), time.Now().AddDate(0, 0, 21))
	if err != nil {
		t.Fatal(err)
	}
	return occurrences
}

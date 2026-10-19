// Package caldav is a small read-only client for the one CalDAV report this
// project needs: an expanded calendar-query over a fixed window. It never
// interprets RRULE; Fastmail's <c:expand> returns concrete instances.
//
// The wire protocol (REPORT, multistatus, basic auth) is handled by
// emersion/go-webdav; this package only builds the query, walks the
// returned events, and maps them onto the domain's Occurrence shape.
package caldav

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	webdav "github.com/emersion/go-webdav"
	gocaldav "github.com/emersion/go-webdav/caldav"
)

// ErrUnauthorized is returned when the server rejects the configured
// credentials. It is kept distinguishable because an app password being
// revoked is the failure that will actually happen.
var ErrUnauthorized = errors.New("caldav: unauthorized")

// Config configures a read-only CalDAV client for one calendar collection.
type Config struct {
	// URL is the calendar collection URL, not a server root. See the trap
	// documented in 090-events-overview.md: PROPFIND /dav/ returns 401 even
	// with valid credentials.
	URL      string
	User     string
	Password string
	// Location resolves floating (VALUE=DATE, no TZID) dates and times to a
	// UTC instant. TZID- and Z-suffixed values use their own explicit zone
	// regardless of this setting.
	Location   *time.Location
	HTTPClient *http.Client
}

// Client issues REPORT requests against one CalDAV calendar collection.
type Client struct {
	inner    *gocaldav.Client
	path     string
	location *time.Location
}

// Occurrence is one concrete calendar instance, already expanded server-side.
type Occurrence struct {
	UID          string
	RecurrenceID string
	Summary      string
	Location     string
	Start        time.Time
	End          time.Time
	AllDay       bool
	Categories   []string // normalised: trimmed and casefolded
}

// New creates a CalDAV client for one calendar collection.
func New(config Config) (*Client, error) {
	if config.URL == "" || config.User == "" || config.Password == "" {
		return nil, errors.New("caldav credentials are incomplete")
	}
	location := config.Location
	if location == nil {
		location = time.UTC
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	// A calendar collection URL without a trailing slash can make some
	// CalDAV servers treat the request target as a different, non-collection
	// resource that does not support REPORT (405 instead of operating on the
	// collection). Normalise rather than trust every caller to remember it.
	rawURL := config.URL
	if !strings.HasSuffix(rawURL, "/") {
		rawURL += "/"
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse caldav URL: %w", err)
	}
	authed := webdav.HTTPClientWithBasicAuth(httpClient, config.User, config.Password)
	inner, err := gocaldav.NewClient(authed, rawURL)
	if err != nil {
		return nil, fmt.Errorf("create caldav client: %w", err)
	}
	// go-webdav v0.7.0's path resolution cleans away a trailing slash when
	// asked to resolve the empty path; pass the collection's own absolute
	// path explicitly (ResolveHref returns an absolute path unmodified) so
	// the REPORT request lands on the collection and not a cousin resource
	// that may answer 405 instead of operating on it.
	return &Client{inner: inner, path: parsed.Path, location: location}, nil
}

// Report fetches every VEVENT instance overlapping [from, to), expanded
// server-side so no RRULE is ever interpreted here.
func (c *Client) Report(ctx context.Context, from, to time.Time) ([]Occurrence, error) {
	query := &gocaldav.CalendarQuery{
		// Expand belongs on the top-level request, not the VEVENT child:
		// go-webdav only reads CalendarCompRequest.Expand at the root
		// (encodeCalendarReq), so an Expand set on a child is silently
		// dropped and the server returns unexpanded RRULE masters whose
		// DTSTART is the series' first instance, not an occurrence.
		CompRequest: gocaldav.CalendarCompRequest{
			Name:   "VCALENDAR",
			Expand: &gocaldav.CalendarExpandRequest{Start: from, End: to},
			Comps: []gocaldav.CalendarCompRequest{
				{Name: ical.CompEvent, AllProps: true},
			},
		},
		CompFilter: gocaldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []gocaldav.CompFilter{
				{Name: ical.CompEvent, Start: from, End: to},
			},
		},
	}

	objects, err := c.inner.QueryCalendar(ctx, c.path, query)
	if err != nil {
		if strings.HasPrefix(err.Error(), fmt.Sprintf("%d ", http.StatusUnauthorized)) {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("caldav report: %w", err)
	}

	var occurrences []Occurrence
	for _, object := range objects {
		if object.Data == nil {
			continue
		}
		for _, event := range object.Data.Events() {
			occurrence, err := occurrenceFromEvent(event, c.location)
			if err != nil {
				return nil, err
			}
			occurrences = append(occurrences, occurrence)
		}
	}
	return occurrences, nil
}

// Put creates one VEVENT on the calendar collection. occ.UID must be unique;
// callers mint it with uuid.NewV7 so retries do not create duplicates.
func (c *Client) Put(ctx context.Context, occ Occurrence) error {
	event := ical.NewEvent()
	event.Props.SetText(ical.PropUID, occ.UID)
	event.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	event.Props.SetText(ical.PropSummary, occ.Summary)
	if occ.Location != "" {
		event.Props.SetText(ical.PropLocation, occ.Location)
	}
	if len(occ.Categories) > 0 {
		prop := ical.NewProp(ical.PropCategories)
		prop.SetTextList(occ.Categories)
		event.Props.Add(prop)
	}
	if occ.AllDay {
		// VALUE=DATE has no timezone in iCalendar. Use the configured
		// calendar location so a UTC instant from GraphQL keeps its intended
		// local calendar date when converted to YYYYMMDD.
		start := occ.Start.In(c.location)
		end := occ.End.In(c.location)
		event.Props.SetDate(ical.PropDateTimeStart, start)
		// DTEND is exclusive for VALUE=DATE; occurrenceFromEvent relies on
		// the ValueType of DTSTART (not DTEND) to detect all-day, so this
		// only needs to be a date one day past the last included day.
		event.Props.SetDate(ical.PropDateTimeEnd, end)
	} else {
		event.Props.SetDateTime(ical.PropDateTimeStart, occ.Start)
		event.Props.SetDateTime(ical.PropDateTimeEnd, occ.End)
	}

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//homie//homie-api//EN")
	cal.Children = append(cal.Children, event.Component)

	path := c.path + occ.UID + ".ics"
	if _, err := c.inner.PutCalendarObject(ctx, path, cal); err != nil {
		if strings.HasPrefix(err.Error(), fmt.Sprintf("%d ", http.StatusUnauthorized)) {
			return ErrUnauthorized
		}
		return fmt.Errorf("caldav put: %w", err)
	}
	return nil
}

func occurrenceFromEvent(event ical.Event, loc *time.Location) (Occurrence, error) {
	uid, err := event.Props.Text(ical.PropUID)
	if err != nil {
		return Occurrence{}, fmt.Errorf("decode UID: %w", err)
	}
	if uid == "" {
		return Occurrence{}, errors.New("VEVENT is missing UID")
	}

	var recurrenceID string
	if prop := event.Props.Get(ical.PropRecurrenceID); prop != nil {
		recurrenceID = prop.Value
	}

	summary, err := event.Props.Text(ical.PropSummary)
	if err != nil {
		return Occurrence{}, fmt.Errorf("VEVENT %q: decode SUMMARY: %w", uid, err)
	}
	location, err := event.Props.Text(ical.PropLocation)
	if err != nil {
		return Occurrence{}, fmt.Errorf("VEVENT %q: decode LOCATION: %w", uid, err)
	}

	var categories []string
	if prop := event.Props.Get(ical.PropCategories); prop != nil {
		values, err := prop.TextList()
		if err != nil {
			return Occurrence{}, fmt.Errorf("VEVENT %q: decode CATEGORIES: %w", uid, err)
		}
		for _, value := range values {
			tag := strings.ToLower(strings.TrimSpace(value))
			if tag != "" {
				categories = append(categories, tag)
			}
		}
	}

	startProp := event.Props.Get(ical.PropDateTimeStart)
	if startProp == nil {
		return Occurrence{}, fmt.Errorf("VEVENT %q is missing DTSTART", uid)
	}
	allDay := startProp.ValueType() == ical.ValueDate

	start, err := event.DateTimeStart(loc)
	if err != nil {
		return Occurrence{}, fmt.Errorf("VEVENT %q: decode DTSTART: %w", uid, err)
	}
	end, err := event.DateTimeEnd(loc)
	if err != nil {
		return Occurrence{}, fmt.Errorf("VEVENT %q: decode DTEND: %w", uid, err)
	}

	return Occurrence{
		UID: uid, RecurrenceID: recurrenceID, Summary: summary, Location: location,
		Start: start, End: end, AllDay: allDay, Categories: categories,
	}, nil
}

<p align="center">
  <img src=".github/homie.svg" alt="Homie" width="150" height="161">
</p>

<h1 align="center">Homie</h1>

<p align="center">A family calendar for the kitchen wall. An open-source alternative to the Skylight Calendar.</p>

Homie runs on a Raspberry Pi behind a touchscreen: the shared calendar, the
weather, and a slow rotation of family photos, all on a screen that lives in the
kitchen. No subscription and no vendor account. It reads your calendar over
CalDAV, keeps everything in SQLite on the device, and the photos never leave
your network.

## What it shows

- **The time and the date**, large enough to read from across the room.
- **Next up and coming up** — a three-week window of events from your CalDAV
  calendar, each tagged with the people it belongs to. Swipe the list up for the
  full window.
- **A photo wall** — photos from a macOS Photos album, captioned with where and
  when they were taken. Tap to move to the next one.
- **Weather** — the current condition and temperature from Apple WeatherKit,
  marked stale if a refresh fails rather than silently showing old numbers.
- **Timers** — start one and it takes over the whole screen.

People and the calendar tags that map onto them are managed from a second web
app at `/manage`, or from the iOS companion app.

## How it fits together

One Go binary serves everything. Both web apps are embedded in it, so a deploy
is a single file copied to the Pi.

```
  macOS                        Raspberry Pi                      iPhone
┌─────────────────┐        ┌───────────────────────────┐      ┌────────────┐
│ PhotoSync.app   │ upload │ homie-api                 │ GQL  │ Homie      │
│ PhotoKit album ─┼───────>│   GraphQL  /graphql       │<────>│ (SwiftUI)  │
└─────────────────┘        │   SQLite + River queue    │      └────────────┘
                           │   display SPA  /          │
  CalDAV ─────────────────>│   manager SPA  /manage    │
  WeatherKit ─────────────>│   photo files  /photos    │
                           └─────────────┬─────────────┘
                                         │ chromium --kiosk under cage, on tty1
                                    touchscreen
```

Calendar and weather refreshes run as periodic background jobs; the Mac pushes
photos in, so the API never calls Apple itself.

## Repository layout

| Path | What it is |
|---|---|
| `services/api` | Go API: GraphQL (gqlgen), SQLite store with goose migrations, River job queue, and the embedded SPAs |
| `apps/display` | The kiosk screen. React, built with Vite |
| `apps/manager` | Admin screen for people and calendar tags, served at `/manage` |
| `apps/mobile/ios` | SwiftUI companion app |
| `apps/photo-sync` | macOS app and LaunchAgent that pushes photos from a Photos album to the API |
| `packages/ui` | Shared React components and design tokens used by both web apps |
| `deploy` | systemd units for the API and the kiosk browser |

## Getting started

You need Docker, [just](https://github.com/casey/just), and — for the Swift
apps — Xcode.

```sh
just dev
```

That brings up the API and both Vite dev servers:

| | URL |
|---|---|
| Display | http://localhost:5173 |
| Manager | http://localhost:5174 |
| GraphQL | http://localhost:3030/graphql |
| Playground | http://localhost:3030/playground |

Calendar and weather need credentials (below) and stay empty without them; the
rest of the screen works regardless.

Run the Go tests with `docker compose run --rm --no-deps api go test ./...`, and
regenerate the GraphQL layer with `just codegen` after editing anything in
`services/api/graph/*.graphqls`.

## Configuration

The API is configured by environment variables. On the Pi they live in
`/opt/homie/api.env`; for local development, in `services/api/.env`.

| Variable | Default | Notes |
|---|---|---|
| `HOMIE_ADDR` | `:3000` | Listen address |
| `HOMIE_LOCAL_ADDR` | *(unset)* | Optional loopback-only listener for the display and manager UI. When enabled, the LAN listener requires `HOMIE_AUTH_TOKEN` (or a generated token at `/var/lib/homie/token`). |
| `HOMIE_AUTH_TOKEN` | *(generated once)* | API bearer token; stored at `/var/lib/homie/token` if unset. The token is logged at startup for client setup. |
| `HOMIE_AUTH_TOKEN_PATH` | `/var/lib/homie/token` | Path used for the persisted bearer token. |
| `HOMIE_PLAYGROUND` | off | Set to `1` to mount `/playground` |
| `HOMIE_DB_PATH` | `/var/lib/homie/homie.db` | Main database |
| `HOMIE_RIVER_DB_PATH` | `/var/lib/homie/river.db` | Job queue database |
| `HOMIE_PHOTOS_PATH` | `/var/lib/homie/photos` | Where uploaded photos are stored |
| `HOMIE_CALDAV_URL`<br>`HOMIE_CALDAV_USER`<br>`HOMIE_CALDAV_PASSWORD` | — | All three or none. Without them, events stay empty |
| `HOMIE_WEATHERKIT_TEAM_ID`<br>`HOMIE_WEATHERKIT_SERVICE_ID`<br>`HOMIE_WEATHERKIT_KEY_ID`<br>`HOMIE_WEATHERKIT_KEY_PATH`<br>`HOMIE_WEATHER_LAT`<br>`HOMIE_WEATHER_LON` | — | All six or none. The key path must be `/var/lib/homie/weatherkit.p8` |

Install the WeatherKit key on the Pi with `just deploy-weatherkit`, which keeps
the private key out of the deploy tree.

On the Pi, get the API token with `sudo journalctl -u homie-api.service -b`; paste
it into the iOS app's Settings and the manager's API token field. Configure Photo
Sync with `defaults write ar.schw.homie.photo-sync HomieAPIToken '<token>'`.

## Photos

Curate the wall by dragging photos into a macOS Photos album named `Homie`.
`apps/photo-sync` reads that album through PhotoKit, reverse-geocodes each
photo's GPS coordinates into a place name, and uploads it to the API. PhotoKit
on a Mac is the only supported way to read an iCloud library unattended and the
only one that still yields location, which is why this piece is a Mac app
rather than something on the Pi. Build it with `just photo-sync`; see
[`apps/photo-sync/README.md`](apps/photo-sync/README.md) for the setup details.

## The Pi

`just deploy` cross-compiles for `linux/arm64` in Docker, copies the binary
over, and restarts both services. The kiosk unit runs Chromium under
[cage](https://github.com/cage-kiosk/cage) on tty1 pointed at the local API; it
restarts on deploy because the web apps are baked into the binary.

While working on the frontend, `just kiosk-dev` points the Pi's browser at the
Vite dev server on your Mac, and `just kiosk-prod` points it back.
When the API is running on the Pi with LAN auth enabled, start the Vite server
with `HOMIE_API_URL=http://homie:3000 HOMIE_AUTH_TOKEN=<token>`. The proxy adds
the token server-side; it is not included in browser code.

## iOS app

`just ios-sim` builds the SwiftUI app and launches it in a simulator. Override
`IOS_SIMULATOR_ID` to pick a different one.

## Status

Homie is a personal project built for one kitchen, and it shows: hostnames and
a few paths are hardcoded for that setup, and it expects a trusted home network.
It is not a packaged product you can install in one step. Expect to read some
code.

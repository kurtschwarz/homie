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
app at `homie.local`, or from the iOS companion app.

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
  CalDAV ─────────────────>│   manager SPA  /          │
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
| `apps/manager` | Admin screen for people and calendar tags, served at `/` on the LAN listener |
| `apps/mobile/ios` | SwiftUI companion app |
| `apps/photo-sync` | macOS app and LaunchAgent that pushes photos from a Photos album to the API |
| `packages/ui` | Shared React components and design tokens used by both web apps |
| `deploy` | systemd units, the blank-cursor theme generator, and Debian packaging sources |

## Getting started

You need Docker, [just](https://github.com/casey/just), and — for the Swift
apps — Xcode.

```sh
just dev
```

That brings up the API and both Vite dev servers, then opens Firefox Developer
Edition with the display and manager in separate tabs. The Firefox DevTools MCP
connects to that browser for local UI verification. Its dedicated profile and
1920x1080 CSS viewport settings live under `.scratch/firefox-dev-profile`.

| | URL |
|---|---|
| Display | http://localhost:5173 |
| Manager | http://localhost:5174/ |
| GraphQL | http://localhost:3030/graphql |
| Playground | http://localhost:3030/playground |

Calendar connections are managed through the GraphQL API. Weather needs the
credentials below. Either feature stays empty when unconfigured; the rest of
the screen works regardless.

Run the Go tests with `docker compose run --rm --no-deps api go test ./...`, and
regenerate the GraphQL layer with `just codegen` after editing anything in
`services/api/graph/*.graphqls`.

## Configuration

The API is configured by environment variables. On the Pi they live in
`/etc/homie/api.env`; for local development, in `services/api/.env`.

| Variable | Default | Notes |
|---|---|---|
| `HOMIE_ADDR` | `:3000` | Listen address. Ignored on the Pi, where `homie-api.socket` binds port 80 and passes it to the service. |
| `HOMIE_LOCAL_ADDR` | *(unset)* | Optional loopback-only listener for the display (kiosk) UI. When enabled, the LAN listener requires a bearer token or a logged-in user session, and the manager is only served there, not on the loopback listener. |
| `HOMIE_AUTH_TOKEN` | *(generated once)* | API bearer token for the display; stored at `/var/lib/homie/token` if unset. The token is logged at startup. It does not unlock the manager - see "Manager access" below. |
| `HOMIE_AUTH_TOKEN_PATH` | `/var/lib/homie/token` | Path used for the persisted bearer token. |
| `HOMIE_SIGNING_SECRET` | *(generated once)* | HMAC key for signed photo URLs, independent of the bearer token so rotating one doesn't invalidate the other's URLs; stored at `/var/lib/homie/signing-secret` if unset. |
| `HOMIE_SIGNING_SECRET_PATH` | `/var/lib/homie/signing-secret` | Path used for the persisted signing secret. |
| `HOMIE_PLAYGROUND` | off | Set to `1` to mount `/playground` |
| `HOMIE_DB_PATH` | `/var/lib/homie/homie.db` | Main database |
| `HOMIE_RIVER_DB_PATH` | `/var/lib/homie/river.db` | Job queue database |
| `HOMIE_PHOTOS_PATH` | `/var/lib/homie/photos` | Where uploaded photos are stored |
| `HOMIE_SECRET_KEY` | *(generated once)* | Base64-encoded 32-byte key used to encrypt calendar credentials; stored at `/var/lib/homie/secret-key` if unset. |
| `HOMIE_SECRET_KEY_PATH` | `/var/lib/homie/secret-key` | Path used for the persisted calendar credential key. |
| `HOMIE_WEATHERKIT_TEAM_ID`<br>`HOMIE_WEATHERKIT_SERVICE_ID`<br>`HOMIE_WEATHERKIT_KEY_ID`<br>`HOMIE_WEATHERKIT_KEY_PATH`<br>`HOMIE_WEATHER_LAT`<br>`HOMIE_WEATHER_LON` | — | All six or none. The key path must be `/var/lib/homie/weatherkit.p8` |

Install the WeatherKit key on the Pi with `just deploy-weatherkit`, which keeps
the private key out of the deploy tree.

On the Pi, get the API token with `sudo journalctl -u homie-api.service -b`; paste
it into the iOS app's Settings. Configure Photo Sync with
`defaults write ar.schw.homie.photo-sync HomieAPIToken '<token>'`.

### Manager access

Opening `homie.local` for the first time shows a setup screen asking for a code,
logged once as its own line by `sudo journalctl -u homie-api.service -b`
(it rotates on every restart until a user account exists). After typing
the code and choosing a username and password, the manager is reached by
logging in from then on - the bearer token above is no longer involved.

If the password is forgotten, drop the user row and reconfigure:

```sh
ssh homie.local "sudo sqlite3 /var/lib/homie/homie.db 'DELETE FROM users'"
ssh homie.local "sudo systemctl restart homie-api.service"
```

The setup code comes back on the next boot. This grants nothing a `root` SSH
session didn't already have.

## Photos

Curate the wall by dragging photos into a macOS Photos album named `Homie`.
`apps/photo-sync` reads that album through PhotoKit, reverse-geocodes each
photo's GPS coordinates into a place name, and uploads it to the API. PhotoKit
on a Mac is the only supported way to read an iCloud library unattended and the
only one that still yields location, which is why this piece is a Mac app
rather than something on the Pi. Build it with `just photo-sync`; see
[`apps/photo-sync/README.md`](apps/photo-sync/README.md) for the setup details.

## The Pi

`just deploy` cross-compiles for `linux/arm64` in Docker, packages the binary,
the systemd units, and the blank-cursor theme into `dist/homie.deb`
(`just package`), ships it to the Pi, and installs it with `apt-get`. The
package's `postinst` restarts both services, which matters because the web
apps are baked into the binary — Chromium needs a restart for frontend
changes to show. The kiosk unit runs Chromium under
[cage](https://github.com/cage-kiosk/cage) on tty1 pointed at the local API.

That first install also names the device. The package depends on
`avahi-daemon`, and `postinst` sets the hostname to `homie`, which together
put the Pi on the network as `homie.local` — the address the justfile and
the rest of this README assume. It happens on a first install only, so
renaming the box later survives upgrades.

On a fresh Pi, the first install has to run by hand, since the sudoers rule
that makes later deploys non-interactive ships inside the package. Reach it
at whatever name it has now, not `homie.local`:

```sh
just package
scp dist/homie.deb raspberrypi.local:/tmp/
ssh -t raspberrypi.local "sudo apt-get install -y /tmp/homie.deb && rm /tmp/homie.deb"
```

After that, `just deploy` installs without a password prompt.

To remove Homie from the Pi, run `sudo sh /usr/share/homie/uninstall.sh`. It
stops the services and deletes the package, config, and data. It leaves the
`homie` user and hostname alone.

### Emoji

Everyone's badge is an emoji, so the kiosk has to draw the same artwork as the
phone that picked it. Nothing in CSS can arrange that on its own: a font stack
names families the machine already has, and a Pi has Noto where a Mac has
Apple's. So `homie-emoji-font.service` fetches
[Apple Color Emoji built for Linux](https://github.com/samuelngs/apple-emoji-ttf)
into `/usr/local/share/fonts` on the first boot after an install, and the
display asks for that family first through the `--font-emoji` token.

The font is downloaded rather than packaged: the artwork is Apple's and is not
ours to redistribute, and at 116 MB it would dwarf the rest of the `.deb`.
Upstream ships a `.deb` of it, but `dpkg` holds its lock through our
`postinst`, so one package cannot install another — hence the bare `.ttf` and a
systemd unit that runs outside the `dpkg` run. The URL and its SHA-256 are
pinned in `deploy/scripts/install-emoji-font.sh`, and re-running it once the
font is installed does nothing. Check on it with
`sudo journalctl -u homie-emoji-font.service -b`.

A 116 MB download over a home link does not always finish, so
`homie-emoji-font.timer` retries every 15 minutes until it does, and a failed
attempt resumes rather than starting over. `Type=oneshot` services cannot use
`Restart=`, which is why the retry is a timer. Both stop once the service
succeeds. The download forces HTTP/1.1: over HTTP/2 this particular file kept
dying mid-transfer with `PROTOCOL_ERROR`, taking twelve minutes to fail where
HTTP/1.1 finishes in about ninety seconds.

Until the download finishes, the kiosk falls back to whatever emoji font the Pi
already had, and Chromium only rereads fonts at startup — so the unit restarts
the kiosk itself once the font lands.

Releases are driven by release-please (`.github/workflows/release-please.yml`).
A push to `main` updates a release PR with the version and changelog from
Conventional Commits; merging it creates a *draft* GitHub release, and the same
workflow then pushes the `homie-vX.Y.Z` tag, because a draft carries no tag of
its own. The two halves never run together: release-please cannot see a draft
release, so the run that cuts one leaves the next release PR to the following
push, which is also the first push with anything to put in it. The tag triggers `.github/workflows/homie-release.yml`, which attaches
the built `.deb` to the draft and publishes it last. The order matters: this
repository has immutable releases turned on, so publishing freezes a release
and GitHub rejects any asset uploaded after that. The same number is stamped
into the API binary.
`just deploy` still builds and installs from your checkout.
Pull requests run only the checks for the area they touch
(`.github/workflows/homie-ci.yml`).

While working on the frontend, `just kiosk-dev` points the Pi's browser at the
Vite dev server on your Mac, and `just kiosk-prod` points it back.
When the API is running on the Pi with LAN auth enabled, start the Vite server
with `HOMIE_API_URL=http://homie.local HOMIE_AUTH_TOKEN=<token>`. The proxy adds
the token server-side; it is not included in browser code.

## iOS app

`just ios-sim` builds the SwiftUI app and launches it in a simulator. Override
`IOS_SIMULATOR_ID` to pick a different one.

## Status

Homie is a personal project built for one kitchen, and it shows: hostnames and
a few paths are hardcoded for that setup, and it expects a trusted home network.
It is not a packaged product you can install in one step. Expect to read some
code.

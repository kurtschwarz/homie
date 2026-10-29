# Photo Sync — PhotoKit upload test

A SwiftUI app that browses the `Homie` album through PhotoKit, previews images, reverse-geocodes GPS metadata, and uploads JPEGs to Homie's GraphQL API using `uploadPhoto`. Uploads use full-size JPEGs and a local version ledger to skip photos already uploaded by this app.

## Build and run

From the repository root:

```sh
just photo-sync
```

This builds the release app at `build/release/Homie Photos.app`. The script builds a minimal app bundle and ad-hoc signs it. Launch through `open`, not by running the executable directly: the Photos privacy prompt must be attributed to the app bundle.

To build and launch a debug app against the local development API on port 3030:

```sh
just dev photo-sync
```

This overwrites the app's saved host with `http://localhost:3030`, so the first local run requires a separate sign-in. To restore the default Pi endpoint, remove the override:

```sh
defaults delete ar.schw.homie.photo-sync HomieAPIURL
defaults delete ar.schw.homie.photo-sync HomieAPIName
defaults delete ar.schw.homie.photo-sync HomieAPIDiscovered
```

## Download a signed build

The **Homie Photos Release** GitHub Actions workflow tests the app. release-please opens a release PR for `apps/photo-sync` from the Conventional Commits that touch it: `feat` and breaking changes bump the minor version (the major version stays 0), and anything else bumps the patch version. Merging that PR creates the `homie-photos-vX.Y.Z` release, and the tag triggers a build that attaches `HomiePhotos.dmg` to it. The app and the disk image are both signed with the Developer ID identity for team `P9DZH3LW75`, and the image is notarized by Apple and stapled so it can be opened normally after download. Pull requests test and build the app with an ad-hoc signature, but do not publish a distribution build.

The workflow requires these repository secrets:

- `MACOS_CERTIFICATE`: the Developer ID Application certificate and private key exported as a password-protected `.p12`, then base64-encoded
- `MACOS_CERTIFICATE_PASSWORD`: the `.p12` export password
- `APPLE_API_KEY_ID`: the App Store Connect API key ID
- `APPLE_API_ISSUER_ID`: the App Store Connect issuer UUID
- `APPLE_API_KEY`: the API key `.p8` file, base64-encoded

To install a build, download `HomiePhotos.dmg` from the [latest release](https://github.com/kurtschwarz/homie/releases/latest) (or directly from `https://github.com/kurtschwarz/homie/releases/latest/download/HomiePhotos.dmg`), open it, and drag **Homie Photos** to Applications. Each run also uploads the image as a workflow artifact, which GitHub retains for 30 days; artifacts are served inside an extra zip wrapper, so prefer the release for a plain download.

The app only adds to the Photos library: it never edits or deletes existing photos, and only adds drawings to the artwork album (see [Drawings](#drawings)). PhotoKit's macOS authorization API uses the `.readWrite` access level. It defaults to the album named `Homie`, shows other albums in the picker, and allows iCloud downloads when an asset is not stored locally. The local ledger is stored in UserDefaults and can be reset by clearing the app's defaults. If the API database is reset, clear the ledger too so photos can be uploaded again.

## Permissions

On launch the app checks three permissions and shows a gate screen over a skeleton sidebar when any is missing: Photos, Local Network, and the Keychain. Each row asks for one permission, and **Continue** stays disabled until all three are allowed. When all three are already granted the screen never appears.

None of the three can be read silently, so asking is what raises the prompt:

- **Photos** is the exception: PhotoKit reports its status without prompting. After a denial macOS will not ask again, so the row offers **Open System Settings**. The status lags the system's own answer by about two seconds, which is why the row shows a spinner after you answer.
- **Local Network** has no status API, and a browser reports `ready` whether or not access was granted. The app advertises its own `_homie-photos-probe._tcp` service and browses for it: seeing itself means access is on, and a policy error means it is off.
- **Keychain** access can only be tested by reading an item, which is what prompts. Launch only checks that the saved item exists, which never prompts.

Local Network and the Keychain therefore remember the last **Allow** rather than re-checking at launch, and that record is tied to the app's code signature: re-signing the app invalidates both grants, so a rebuilt app asks again instead of showing a stale tick. Revoking either in System Settings is only noticed the next time the row is used. **Reset saved sign-in** deletes the stored session item so the next sign-in creates a fresh one.

Choose **Homie Photos → Reset Homie Photos…** to return the app to onboarding. The reset removes saved hosts, sessions, upload history, and cached places, but macOS keeps the Photos permission. It also removes the local API override set by `just dev photo-sync`; run that recipe again to restore the development endpoint.

## Background sync

Homie Photos runs from the menu bar with no Dock icon. While it runs it syncs every 5 minutes without a window, so drawings reach the artwork album on their own. Sync is skipped until the permissions check has run, you are signed in, Photos access is granted, and albums are chosen; the menu then reads "Sign in needed" or "Open Homie Photos to finish setup". Two syncs never overlap: the timer and **Sync Now** both skip while one is running. Turning off "keep syncing" stops the timer.

The menu bar shows the last sync time or current progress, and **Open Homie Photos…** opens the window for album selection and sign-in. The window is a fixed 680 × 760, always-dark glass window with a sidebar; it needs macOS 15 or later, and uses Liquid Glass on macOS 26. **Start at Login** registers the app as a login item. Login items need a signed build in Applications; macOS may refuse for an unsigned local build.

## Drawings

Each sync first adds drawings saved on the display to the artwork album, then uploads photos. For every drawing Photo Sync downloads the JPEG, creates a Photos asset from those bytes with the drawn-at time as its creation date, adds it to the artwork album, and only then tells Homie it was pushed. Pushed assets are skipped when photos are uploaded, so an artwork album that is also synced does not send drawings back. Nothing is pushed until an artwork album is chosen.

What survives the import into Photos, checked on a real drawing:

| Field                   | Result                                            |
| ----------------------- | ------------------------------------------------- |
| Creation date           | Kept; shown as the drawn-at time                  |
| `ImageDescription`      | Kept; shown as the caption                        |
| Camera make and model   | Kept; shown as "Homie Display"                    |
| GPS                     | Not checked; the drawing tested had no location   |
| `Artist`                | Not checked; present in the file, but Photos' Info panel has no field for it |

## API endpoint

Uploads go to `http://homie/graphql`, the Pi, by default. `just dev photo-sync` configures the local endpoint automatically. To point the app at a local API manually instead:

```sh
defaults write ar.schw.homie.photo-sync HomieAPIURL http://localhost:3030
defaults delete ar.schw.homie.photo-sync HomieAPIURL  # back to the Pi
```

A bare origin gets `/graphql` appended, so the value is spelled the same way as `HOMIE_API_URL` for the web apps. `HOMIE_API_URL` is also read, but only when the executable is run from a shell: a bundle launched by `open` does not inherit the environment. For a local API, start it with `docker compose up api`, which publishes port 3030.

Sign in with the Homie admin username and password. Photo Sync stores its session in the macOS Keychain, and the session appears in the manager's Active Sessions list, where it can be revoked. Photo Sync asks you to sign in again when the session expires or is revoked.

Photo Sync discovers Homie devices advertised over Bonjour and also checks `homie.local` as a fallback. Use **Change** on the sign-in screen, or the display button after signing in, to select another discovered device or enter an address manually. Sessions are stored separately for each address.

The bundle declares two Bonjour service types: `_homie._tcp` to find Homie devices, and `_homie-photos-probe._tcp`, which the app advertises only to test its own Local Network access (see [Permissions](#permissions)). The app bundle sets `NSAllowsLocalNetworking` because the API is plain HTTP. App Transport Security exempts `localhost` but not the Pi's hostname.

The server accepts JPEG uploads up to 20 MB. Photo Sync encodes full-resolution images and retries oversized images at lower JPEG quality; images that still exceed the server limit are marked as failed.

## Reset permissions

If authorization gets stuck after rebuilding, reset the app's TCC grants and reopen it:

```sh
tccutil reset Photos ar.schw.homie.photo-sync
tccutil reset All ar.schw.homie.photo-sync  # Photos and Local Network together
```

`tccutil` does not clear the app's own record of the last **Allow** for Local Network and the Keychain. Those records are tied to the code signature, so a rebuild clears them; to clear them without rebuilding, remove them directly:

```sh
defaults delete ar.schw.homie.photo-sync localNetworkGrantedForCode
defaults delete ar.schw.homie.photo-sync keychainGrantedForCode
```

Local builds use an ad-hoc signature. The GitHub Actions build uses Developer ID signing and Apple notarization for distribution.

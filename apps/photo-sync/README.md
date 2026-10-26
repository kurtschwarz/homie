# Photo Sync — PhotoKit upload test

A SwiftUI app that browses the `Homie` album through PhotoKit, previews images, reverse-geocodes GPS metadata, and uploads JPEGs to Homie's GraphQL API using `uploadPhoto`. Uploads use full-size JPEGs and a local version ledger to skip photos already uploaded by this app.

## Build and run

From the repository root:

```sh
just photo-sync
```

The script builds a minimal `PhotoSync.app` bundle and ad-hoc signs it. Launch through `open`, not by running the executable directly: the Photos privacy prompt must be attributed to the app bundle.

The app does not modify the Photos library. PhotoKit's macOS authorization API uses the `.readWrite` access level. It defaults to the album named `Homie`, shows other albums in the picker, and allows iCloud downloads when an asset is not stored locally. The local ledger is stored in UserDefaults and can be reset by clearing the app's defaults. If the API database is reset, clear the ledger too so photos can be uploaded again.

## API endpoint

Uploads go to `http://homie:3000/graphql`, the Pi, by default. To point the app at a local API instead:

```sh
defaults write ar.schw.homie.photo-sync HomieAPIURL http://localhost:3030
defaults delete ar.schw.homie.photo-sync HomieAPIURL  # back to the Pi
```

A bare origin gets `/graphql` appended, so the value is spelled the same way as `HOMIE_API_URL` for the web apps. `HOMIE_API_URL` is also read, but only when the executable is run from a shell: a bundle launched by `open` does not inherit the environment. For a local API, start it with `docker compose up api`, which publishes port 3030.

Sign in with the Homie admin username and password. Photo Sync stores its session in the macOS Keychain, and the session appears in the manager's Active Sessions list, where it can be revoked. Photo Sync asks you to sign in again when the session expires or is revoked.

Photo Sync discovers Homie devices advertised over Bonjour and also checks `homie.local:3000` as a fallback. Use **Change** on the sign-in screen, or the display button after signing in, to select another discovered device or enter an address manually. Sessions are stored separately for each address.

The app bundle sets `NSAllowsLocalNetworking` because the API is plain HTTP. App Transport Security exempts `localhost` but not the Pi's hostname.

The server accepts JPEG uploads up to 20 MB. Photo Sync encodes full-resolution images and retries oversized images at lower JPEG quality; images that still exceed the server limit are marked as failed.

## Reset Photos permission

If authorization gets stuck after rebuilding, reset the app's TCC grant and reopen it:

```sh
tccutil reset Photos ar.schw.homie.photo-sync
```

Ad-hoc signatures are suitable for this local spike; they are not a distribution/signing strategy.

# Photo Sync — PhotoKit upload test

A SwiftUI app that browses the `Homie` album through PhotoKit, previews images, reverse-geocodes GPS metadata, and uploads JPEGs to Homie's GraphQL API using `uploadPhoto`. Uploads use full-size JPEGs and a local version ledger to skip photos already uploaded by this app.

## Build and run

From the repository root:

```sh
just photo-sync
```

The script builds a minimal `PhotoSync.app` bundle and ad-hoc signs it. Launch through `open`, not by running the executable directly: the Photos privacy prompt must be attributed to the app bundle.

The app does not modify the Photos library. PhotoKit's macOS authorization API uses the `.readWrite` access level. It defaults to the album named `Homie`, shows other albums in the picker, and allows iCloud downloads when an asset is not stored locally. The API endpoint is currently hardcoded to `http://localhost:3030/graphql`; start the API locally with `docker compose up api` before uploading. The local ledger is stored in UserDefaults and can be reset by clearing the app's defaults. If the API database is reset, clear the ledger too so photos can be uploaded again.

The server accepts JPEG uploads up to 20 MB. Photo Sync encodes full-resolution images and retries oversized images at lower JPEG quality; images that still exceed the server limit are marked as failed.

## Reset Photos permission

If authorization gets stuck after rebuilding, reset the app's TCC grant and reopen it:

```sh
tccutil reset Photos ar.schw.homie.photo-sync
```

Ad-hoc signatures are suitable for this local spike; they are not a distribution/signing strategy.

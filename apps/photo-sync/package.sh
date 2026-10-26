#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
APP="$ROOT/build/Homie Photos.app"
DMG="$ROOT/build/HomiePhotos.dmg"
STAGE="$ROOT/build/dmg"

if [[ ! -d "$APP" ]]; then
    printf 'No app at %s. Run build.sh first.\n' "$APP" >&2
    exit 1
fi

rm -rf "$STAGE"
rm -f "$DMG"
mkdir -p "$STAGE"

# ditto rather than cp, to keep the bundle's signature intact.
ditto "$APP" "$STAGE/Homie Photos.app"
ln -s /Applications "$STAGE/Applications"

hdiutil create \
    -volname 'Homie Photos' \
    -srcfolder "$STAGE" \
    -fs HFS+ \
    -format UDZO \
    -ov \
    "$DMG"

rm -rf "$STAGE"

if [[ -n "${CODESIGN_IDENTITY:-}" ]]; then
    codesign --force --sign "$CODESIGN_IDENTITY" --timestamp "$DMG"
    codesign --verify --strict --verbose=2 "$DMG"
fi

printf 'Packaged %s\n' "$DMG"

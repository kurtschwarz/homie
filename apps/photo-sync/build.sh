#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
APP="$ROOT/build/Homie Photos.app"

swift build --package-path "$ROOT" -c release
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$ROOT/.build/release/PhotoSync" "$APP/Contents/MacOS/PhotoSync"
cp "$ROOT/Info.plist" "$APP/Contents/Info.plist"
cp "$ROOT/Resources/AppIcon.icns" "$APP/Contents/Resources/AppIcon.icns"

if [[ -n "${CODESIGN_IDENTITY:-}" ]]; then
    codesign \
        --force \
        --sign "$CODESIGN_IDENTITY" \
        --timestamp \
        --options runtime \
        --entitlements "$ROOT/PhotoSync.entitlements" \
        "$APP"
else
    codesign --force --sign - "$APP"
fi

codesign --verify --deep --strict --verbose=2 "$APP"
printf 'Built %s\n' "$APP"

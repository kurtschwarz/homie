#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
CONFIGURATION="${1:-release}"
APP="$ROOT/build/$CONFIGURATION/Homie Photos.app"

swift build --package-path "$ROOT" -c "$CONFIGURATION"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp "$ROOT/.build/$CONFIGURATION/PhotoSync" "$APP/Contents/MacOS/PhotoSync"
cp "$ROOT/Info.plist" "$APP/Contents/Info.plist"

# Must happen before codesign, which seals Info.plist.
if [[ -n "${VERSION:-}" ]]; then
    plutil -replace CFBundleShortVersionString -string "$VERSION" "$APP/Contents/Info.plist"
    plutil -replace CFBundleVersion -string "$(git -C "$ROOT" rev-list --count HEAD)" "$APP/Contents/Info.plist"
fi

if commit="$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null)"; then
    plutil -replace HomieGitCommit -string "$commit" "$APP/Contents/Info.plist"
fi
cp "$ROOT/Resources/AppIcon.icns" "$APP/Contents/Resources/AppIcon.icns"
cp "$ROOT"/Resources/HomieMenuTemplate*.png "$APP/Contents/Resources/"

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

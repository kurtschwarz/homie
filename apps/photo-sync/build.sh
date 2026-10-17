#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
APP="$ROOT/build/PhotoSync.app"

swift build --package-path "$ROOT" -c release
mkdir -p "$APP/Contents/MacOS"
cp "$ROOT/.build/release/PhotoSync" "$APP/Contents/MacOS/PhotoSync"
cp "$ROOT/Info.plist" "$APP/Contents/Info.plist"
codesign --force --sign - "$APP"
printf 'Built %s\n' "$APP"

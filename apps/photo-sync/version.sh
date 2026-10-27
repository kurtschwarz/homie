#!/bin/bash
set -euo pipefail

# Prints the next Photo Sync version, or with "notes" the changelog since the
# last release. See scripts/version.sh for the versioning rules.

ROOT="$(cd "$(dirname "$0")" && pwd)"
MODE="${1:-version}"

base="0.0.0"
if [[ "$MODE" != "notes" ]]; then
    base="$(plutil -extract CFBundleShortVersionString raw "$ROOT/Info.plist")"
fi

exec "$ROOT/../../scripts/version.sh" photo-sync-v "$base" "$MODE" apps/photo-sync

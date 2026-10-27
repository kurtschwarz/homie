#!/bin/bash
set -euo pipefail

# Prints the next version, or with "notes" the changelog since the last release.
# The version comes from Conventional Commits that touch apps/photo-sync since
# the newest photo-sync-v* tag. While the major version is 0, breaking changes
# bump the minor version.

ROOT="$(cd "$(dirname "$0")" && pwd)"
REPO="$(git -C "$ROOT" rev-parse --show-toplevel)"
SCOPE='apps/photo-sync'
PREFIX='photo-sync-v'

last_tag="$(git -C "$REPO" tag --list "${PREFIX}*" --sort=-v:refname | sed -n 1p)"

if [[ -n "$last_tag" ]]; then
    range="$last_tag..HEAD"
else
    range="HEAD"
fi

if [[ "${1:-}" == "notes" ]]; then
    git -C "$REPO" log --no-merges --format='- %s (%h)' "$range" -- "$SCOPE"
    exit 0
fi

if [[ -n "$last_tag" ]]; then
    base="${last_tag#"$PREFIX"}"
else
    base="$(plutil -extract CFBundleShortVersionString raw "$ROOT/Info.plist")"
fi

IFS=. read -r major minor patch <<< "$base"

level=0
while IFS= read -r subject; do
    if [[ "$subject" =~ ^[a-z]+(\([^\)]*\))?!: ]]; then
        level=2
    elif [[ "$subject" =~ ^feat(\([^\)]*\))?: ]] && [[ "$level" -lt 1 ]]; then
        level=1
    fi
done < <(git -C "$REPO" log --no-merges --format='%s' "$range" -- "$SCOPE")

if git -C "$REPO" log --no-merges --format='%B' "$range" -- "$SCOPE" | grep -q '^BREAKING CHANGE'; then
    level=2
fi

if [[ "$major" -eq 0 && "$level" -eq 2 ]]; then
    level=1
fi

case "$level" in
    2) major=$((major + 1)); minor=0; patch=0 ;;
    1) minor=$((minor + 1)); patch=0 ;;
    *) patch=$((patch + 1)) ;;
esac

printf '%s.%s.%s\n' "$major" "$minor" "$patch"

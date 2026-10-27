#!/bin/bash
set -euo pipefail

# Usage: version.sh <tag-prefix> <fallback-base> <version|notes> <path>...
#
# Prints the next version, or with "notes" the changelog since the last
# release. The version comes from Conventional Commits that touch the given
# paths since the newest <tag-prefix>* tag. While the major version is 0,
# breaking changes bump the minor version. <fallback-base> is the version to
# bump when no tag exists yet.

if [[ $# -lt 4 ]]; then
    printf 'usage: %s <tag-prefix> <fallback-base> <version|notes> <path>...\n' "$0" >&2
    exit 2
fi

PREFIX="$1"
FALLBACK_BASE="$2"
MODE="$3"
shift 3
SCOPES=("$@")

REPO="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

last_tag="$(git -C "$REPO" tag --list "${PREFIX}*" --sort=-v:refname | sed -n 1p)"

if [[ -n "$last_tag" ]]; then
    range="$last_tag..HEAD"
else
    range="HEAD"
fi

if [[ "$MODE" == "notes" ]]; then
    git -C "$REPO" log --no-merges --format='- %s (%h)' "$range" -- "${SCOPES[@]}"
    exit 0
fi

if [[ -n "$last_tag" ]]; then
    base="${last_tag#"$PREFIX"}"
else
    base="$FALLBACK_BASE"
fi

IFS=. read -r major minor patch <<< "$base"

level=0
while IFS= read -r subject; do
    if [[ "$subject" =~ ^[a-z]+(\([^\)]*\))?!: ]]; then
        level=2
    elif [[ "$subject" =~ ^feat(\([^\)]*\))?: ]] && [[ "$level" -lt 1 ]]; then
        level=1
    fi
done < <(git -C "$REPO" log --no-merges --format='%s' "$range" -- "${SCOPES[@]}")

if git -C "$REPO" log --no-merges --format='%B' "$range" -- "${SCOPES[@]}" | grep -q '^BREAKING CHANGE'; then
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

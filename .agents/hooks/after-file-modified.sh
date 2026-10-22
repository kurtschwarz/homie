#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -gt 0 ]; then
  file_path="$1"
else
  file_path="$(jq -r '.tool_input.file_path // empty')"
fi

[ -n "$file_path" ] || exit 0

case "$file_path" in
  /*) ;;
  *) file_path="$PWD/$file_path" ;;
esac

case "$file_path" in
  *.ts|*.tsx|*.js|*.jsx|*.mjs|*.cjs|*.mts|*.cts|*.css|*.json|*.jsonc) ;;
  *) exit 0 ;;
esac

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
cd "$repo_root"

pnpm exec biome check --write --linter-enabled=false --no-errors-on-unmatched "$file_path"

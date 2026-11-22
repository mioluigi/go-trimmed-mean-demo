#!/usr/bin/env bash
# review-bundle.sh: packs review evidence into ONE text file (SFEDD §5.4), because separate
# uploads to a reviewer can be dropped silently.
#
# Usage: scripts/review-bundle.sh OUTPUT FILE...
# Each file is preceded by a "===== FILE: path =====" header; the bundle starts with a list of them.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "review-bundle: $*" >&2; exit 1; }

[ "$#" -ge 2 ] || fail "usage: scripts/review-bundle.sh OUTPUT FILE..."
output="$1"; shift
for file in "$@"; do
  [ -f "$file" ] || fail "missing file: $file"
done

{
  echo "REVIEW BUNDLE: $# files, from commit $(git rev-parse --short HEAD)"
  for file in "$@"; do echo "  - $file"; done
  for file in "$@"; do
    printf '\n===== FILE: %s =====\n' "$file"
    cat "$file"
  done
  printf '\n===== END OF BUNDLE (%s files) =====\n' "$#"
} > "$output"
echo "review-bundle: wrote $output ($# files, $(wc -l < "$output" | tr -d ' ') lines)"

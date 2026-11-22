#!/usr/bin/env bash
# check-inputs.sh: the data gate (SFEDD §5.2; constitution 8.1).
#
# Checks every sample against data/MANIFEST.tsv: present, expected count, recorded size and SHA-256,
# header "value", n observations (at least 100), every value well formed for its kind (whole numbers
# for "integer", exactly 4 decimals for "double", as Week 8), and no stray files. Run before committing data, and
# by measure.sh and acceptance.sh.
#
# Environment: DATA_DIR (default data), EXPECTED_FILES (default 6).
set -euo pipefail
cd "$(dirname "$0")/.."
data="${DATA_DIR:-data}"
expected="${EXPECTED_FILES:-6}"
manifest="$data/MANIFEST.tsv"
problems=0
problem() { echo "check-inputs: $*" >&2; problems=$((problems + 1)); }

# sha256_of FILE: prints the SHA-256 with whichever tool this system has.
sha256_of() {
  if command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
  else sha256sum "$1" | awk '{print $1}'; fi
}

[ -f "$manifest" ] || { echo "check-inputs: $manifest is missing; run scripts/capture-r.sh on the Mac." >&2; exit 1; }

integer_pattern='^-?[0-9]+$'
double_pattern='^-?[0-9]+\.[0-9]{4}$'

count=0
while IFS=$'\t' read -r file kind n generator seed bytes sha256; do
  case "$file" in ''|'#'*|file) continue ;; esac
  count=$((count + 1))
  path="$data/$file"
  if [ ! -f "$path" ]; then problem "$file is listed in the manifest but missing"; continue; fi
  [ "$(wc -c < "$path" | tr -d ' ')" = "$bytes" ] || problem "$file: size differs from the manifest"
  [ "$(sha256_of "$path")" = "$sha256" ] || problem "$file: SHA-256 differs from the manifest"
  [ "$(head -n 1 "$path")" = "value" ] || problem "$file: first line is not the header \"value\""
  observations=$(( $(wc -l < "$path" | tr -d ' ') - 1 ))
  [ "$observations" = "$n" ] || problem "$file: $observations observations, manifest says $n"
  [ "$observations" -ge 100 ] || problem "$file: fewer than 100 observations"
  case "$kind" in
    integer) pattern="$integer_pattern" ;;
    double) pattern="$double_pattern" ;;
    *) problem "$file: unknown kind \"$kind\""; continue ;;
  esac
  malformed=$(tail -n +2 "$path" | grep -c -v -E "$pattern" || true)
  [ "$malformed" = "0" ] || problem "$file: $malformed lines are not well-formed $kind values"
done < "$manifest"
[ "$count" = "$expected" ] || problem "manifest lists $count samples, expected $expected"

for path in "$data"/* "$data"/.[!.]*; do
  [ -e "$path" ] || continue
  name="$(basename "$path")"
  [ "$name" = "MANIFEST.tsv" ] && continue
  grep -q "^$name	" "$manifest" || problem "stray file in $data: $name"
done

if [ "$problems" -gt 0 ]; then
  echo "check-inputs: FAILED with $problems problem(s)" >&2
  exit 1
fi
echo "check-inputs: ok ($count samples match $manifest)"

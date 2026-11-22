#!/usr/bin/env bash
# capture-r.sh: runs r/capture-reference.R on the Mac, then writes data/MANIFEST.tsv and runs the data gate.
#
# The samples and R's reference results are generated once, by this committed, seeded script
# (constitution 8.3), so R and Go provably read the same bytes.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "capture-r: $*" >&2; exit 1; }
command -v Rscript >/dev/null 2>&1 || fail "Rscript is not installed; this step runs on the Mac with R."

# sha256_of FILE: prints the SHA-256 with whichever tool this system has.
sha256_of() {
  if command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
  else sha256sum "$1" | awk '{print $1}'; fi
}

rm -f data/*.csv
Rscript r/capture-reference.R

seed="$(awk -F': ' '/^seed:/ {print $2}' testdata/r-reference/r-session.txt)"
{
  printf 'file\tkind\tn\tgenerator\tseed\tbytes\tsha256\n'
  for path in data/*.csv; do
    name="$(basename "$path")"
    case "$name" in integers-*) kind=integer ;; *) kind=double ;; esac
    n=$(( $(wc -l < "$path" | tr -d ' ') - 1 ))
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$name" "$kind" "$n" "r/capture-reference.R" "$seed" \
      "$(wc -c < "$path" | tr -d ' ')" "$(sha256_of "$path")"
  done
} > data/MANIFEST.tsv

scripts/check-inputs.sh

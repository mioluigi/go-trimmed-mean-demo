#!/usr/bin/env bash
# capture-bootstrap-r.sh COURSE_FILE: runs r/capture-bootstrap-reference.R on the Mac with Luis's copy of
# the course file run-bootstrap-robust.R (unchanged, never committed), and writes a manifest with the
# SHA-256 of every output and of the course file, so the reference names exactly what produced it.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "capture-bootstrap-r: $*" >&2; exit 1; }
[ "$#" -eq 1 ] || fail "usage: scripts/capture-bootstrap-r.sh /path/to/run-bootstrap-robust.R"
course="$1"
[ -f "$course" ] || fail "course file not found: $course"
command -v Rscript >/dev/null 2>&1 || fail "Rscript is not installed; this step runs on the Mac with R."
case "$(cd "$(dirname "$course")" && pwd)" in "$(pwd)"*) fail "keep the course file outside this repository (it is not committed)." ;; esac

sha256_of() {
  if command -v shasum >/dev/null 2>&1; then shasum -a 256 "$1" | awk '{print $1}'
  else sha256sum "$1" | awk '{print $1}'; fi
}

out=testdata/bootstrap-reference
rm -rf "$out"
Rscript r/capture-bootstrap-reference.R "$course"
{
  printf 'file\tbytes\tsha256\n'
  printf 'course file run-bootstrap-robust.R (not committed)\t%s\t%s\n' "$(wc -c < "$course" | tr -d ' ')" "$(sha256_of "$course")"
  for path in "$out"/*; do
    printf '%s\t%s\t%s\n' "$(basename "$path")" "$(wc -c < "$path" | tr -d ' ')" "$(sha256_of "$path")"
  done
} > "$out/MANIFEST.tsv"
cat "$out/MANIFEST.tsv"

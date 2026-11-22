#!/usr/bin/env bash
# acceptance.sh: checks the acceptance criteria of specs/001-trimmed-mean/spec.md (§9) on FRESH CLONES of
# both repositories from GitHub, in a temporary folder, with an empty Go module cache. It changes nothing
# in this copy. Run it on the Mac (it needs GitHub and the Go module proxy) as the last act before
# submitting.
#
# Environment: PACKAGE_VERSION (default: the version the program's go.mod requires),
#              DEMO_REPO, PACKAGE_REPO (default: the two GitHub repositories).
# Exit status: 0 if no check failed, 1 otherwise.
set -uo pipefail
cd "$(dirname "$0")/.."
demo_repo="${DEMO_REPO:-https://github.com/mioluigi/go-trimmed-mean-demo.git}"
package_repo="${PACKAGE_REPO:-https://github.com/mioluigi/trimmedmean.git}"
version="${PACKAGE_VERSION:-$(awk '$1 == "require" && $2 == "github.com/mioluigi/trimmedmean" {print $3}' go.mod)}"

work="$(mktemp -d)"
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT
export GOMODCACHE="$work/modcache" GOFLAGS=-mod=readonly
passed=0; failed=0

# check NAME COMMAND...: runs the command quietly; PASS if it succeeds.
check() {
  local name="$1"; shift
  if "$@" > "$work/output.txt" 2>&1; then
    echo "PASS  $name"; passed=$((passed + 1))
  else
    echo "FAIL  $name"; sed 's/^/        /' "$work/output.txt" | tail -n 8; failed=$((failed + 1))
  fi
}
# exits CODE COMMAND...: succeeds if the command exits with exactly that code.
exits() { local want="$1"; shift; "$@" > /dev/null 2>&1; [ "$?" -eq "$want" ]; }
# says TEXT COMMAND...: succeeds if the command's output contains the text.
says() { local text="$1"; shift; "$@" > "$work/said.txt" 2>&1; grep -q -F -- "$text" "$work/said.txt"; }
# has FILE TEXT...: succeeds if the file contains every text.
has() { local file="$1"; shift; for text in "$@"; do grep -q -F -- "$text" "$file" || { echo "$file lacks: $text"; return 1; }; done; }
# offline COMMAND...: runs with no module proxy and with an Rscript that always fails first on PATH,
# so a test that needed the internet or R would fail (spec NFR-6).
mkdir -p "$work/no-r"; printf '#!/bin/sh\necho "Rscript is not available in acceptance" >&2\nexit 127\n' > "$work/no-r/Rscript"
chmod +x "$work/no-r/Rscript"
offline() { GOPROXY=off PATH="$work/no-r:$PATH" "$@"; }
no_replace() { ! grep -q -E '^[[:space:]]*replace' go.mod && [ ! -e go.work ]; }

echo "== Fresh clones (package $version)"
git clone -q --branch "$version" "$package_repo" "$work/trimmedmean" 2>&1 || { echo "FAIL  package tag $version cannot be cloned"; exit 1; }
git clone -q "$demo_repo" "$work/demo" || { echo "FAIL  program repository cannot be cloned"; exit 1; }
echo "PASS  both repositories cloned; package at tag $version"; passed=$((passed + 1))

echo "== Package github.com/mioluigi/trimmedmean@$version"
cd "$work/trimmedmean"
check "go.mod has no require and no replace; no go.work (FR-R1, AC-7, AC-8)" sh -c '! grep -q -E "^[[:space:]]*(require|replace)" go.mod && [ ! -e go.work ]'
check "gofmt reports no file; go vet is clean (AC-16)" sh -c '[ -z "$(gofmt -l .)" ] && go vet ./...'
check "all package tests pass offline, without R: hand-computed, R reference, examples, counts, quality (AC-1..7, 16)" offline go test -count=1 ./...
check "the R reference test ran and passed on this machine (AC-3)" says "--- PASS: TestMeanAgreesWithR" offline go test -count=1 -v -run 'AgreesWithR' .
check "package README: install, use, errors, limits, differences from R, tests, licence (FR-R4, AC-13)" \
  has README.md "go get github.com/mioluigi/trimmedmean@" "## Install" "## Use" "## Errors" "## Limits" "## Differences from R" "## Tests" "## Licence"
check "LICENSE present" test -s LICENSE

echo "== Program github.com/mioluigi/go-trimmed-mean-demo (fresh clone of main)"
cd "$work/demo"
check "go.mod requires trimmedmean $version; no replace; no go.work (FR-7, AC-8)" sh -c "grep -q 'github.com/mioluigi/trimmedmean $version' go.mod && ! grep -q -E '^[[:space:]]*replace' go.mod && [ ! -e go.work ]"
check "builds from an empty module cache with -mod=readonly, fetching the tag (AC-8)" go build -o "$work/trimmed-mean-demo" .
check "gofmt reports no file; go vet is clean (NFR-1)" sh -c '[ -z "$(gofmt -l .)" ] && go vet ./...'
check "all program tests pass offline, without R: built-binary exit codes and messages (AC-9, AC-10)" offline go test -count=1 ./...
check "samples pass the data gate: 6 files, at least 100 values each (AC-3, S6)" scripts/check-inputs.sh
check "the data gate fails on a tampered sample (AC-15)" sh -c 'cp -R data "$0/data" && echo 1.23456 >> "$0/data/floats-100.csv" && ! DATA_DIR="$0/data" scripts/check-inputs.sh' "$work"
check "no arguments: exit 0 and 84 of 84 comparisons agree (AC-9)" says "84 of 84 comparisons with R agree" "$work/trimmed-mean-demo" -out "$work/comparison.tsv"
check "one file, asymmetric: exit 0 (FR-9)" exits 0 "$work/trimmed-mean-demo" -data data/floats-100.csv -trim 0.05,0.10
check "trim 0.6: exit 2 (AC-10)" exits 2 "$work/trimmed-mean-demo" -data data/floats-100.csv -trim 0.6
check "build scripts make the macOS binaries, the Windows .exe and the .app (AC-12)" sh -c \
  'DIST="$0/dist" scripts/build.sh && DIST="$0/dist" scripts/build-macos-app.sh && test -s "$0/dist/trimmed-mean-demo.exe" -a -s "$0/dist/trimmed-mean-demo-macos-arm64" -a -x "$0/dist/TrimmedMean.app/Contents/MacOS/TrimmedMean"' "$work"
check "README's generated blocks match results/ (AC-14)" scripts/readme.sh --check
check "results/comparison.tsv: every row PASS" sh -c '[ "$(tail -n +2 results/comparison.tsv | grep -c -v "	PASS$")" = 0 ] && [ "$(tail -n +2 results/comparison.tsv | wc -l | tr -d " ")" = 84 ]'
measured="$(awk '/^commit:/ {print $2}' results/machine.txt 2>/dev/null)"
check "what was measured is what is submitted: no code or data change since commit ${measured:-?}" \
  sh -c "[ -n '$measured' ] && git diff --quiet '$measured' HEAD -- '*.go' go.mod go.sum data testdata"
check "results were measured with trimmedmean $version" has results/machine.txt "trimmedmean: $version"
check "program README: get and use the package, build, run, test, results, limits, GenAI Tools (FR-13, AC-13)" \
  has README.md "go get github.com/mioluigi/trimmedmean@$version" "## Getting and using the package" "## Building" "## Running" "## Testing" "## Results" "## Limits" "## GenAI Tools"
check "AI dialogs are indexed: docs/ai-dialogs/README.md lists every file there (AC-13)" sh -c \
  'test -s docs/ai-dialogs/README.md && for f in docs/ai-dialogs/*; do n="$(basename "$f")"; [ "$n" = README.md ] || grep -q -F "$n" docs/ai-dialogs/README.md || { echo "not indexed: $n"; exit 1; }; done'

echo
echo "acceptance: $passed passed, $failed failed"
[ "$failed" -eq 0 ]

#!/usr/bin/env bash
# measure.sh: the official run on Luis's Mac (SFEDD §5.3; spec FR-8, NFR-2, NFR-3, NFR-4).
#
# Close other programs first. The script:
#   0. refuses a stale or modified copy (check-fresh.sh) or failing samples (check-inputs.sh);
#   1. runs the comparison with R (correctness first; nothing else runs if it fails)
#        -> results/comparison.tsv, comparison-run.txt (what the user sees), comparison-log.txt;
#   2. runs the package benchmarks, 5 times each, -benchmem -> results/benchmarks.txt;
#   3. profiles the program on the largest committed sample and on a 1,000,000-value file generated
#      here (not committed: a profile of 1,000 values is too short to sample)
#        -> results/cpu-profile.txt, mem-profile.txt (and the .prof files);
#   4. records the machine, Go version, package version and commit -> results/machine.txt;
#   5. regenerates results/tables.md and the README's generated blocks (scripts/readme.sh).
# It records no user name, host name or home path.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "measure: $*" >&2; exit 1; }
[ "$(uname -s)" = "Darwin" ] || fail "official measurements run on the Mac (this is $(uname -s))."
scripts/check-fresh.sh
scripts/check-inputs.sh

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p results
go build -o "$work/trimmed-mean-demo" .

echo "measure: step 1 of 5: comparison with R"
"$work/trimmed-mean-demo" -out results/comparison.tsv > results/comparison-run.txt 2> "$work/log.txt" \
  || { cat results/comparison-run.txt "$work/log.txt"; fail "the comparison with R failed; nothing else was run."; }
sed -E 's/^time=[^ ]+ //' "$work/log.txt" > results/comparison-log.txt   # drop timestamps: reproducible
cat results/comparison-run.txt

echo "measure: step 2 of 5: package benchmarks (5 runs each)"
package_version="$(go list -m -f '{{.Version}}' github.com/mioluigi/trimmedmean)"
go test -run '^$' -bench . -benchmem -count 5 github.com/mioluigi/trimmedmean > results/benchmarks.txt

echo "measure: step 3 of 5: profiles"
largest="$(ls -S data/*.csv | head -n 1)"
awk 'BEGIN { srand(431); print "value"; for (i = 0; i < 1000000; i++) printf "%.4f\n", (rand() - 0.5) * 100 }' > "$work/million.csv"
"$work/trimmed-mean-demo" -data "$largest" -cpuprofile results/cpu-largest-sample.prof > /dev/null 2>&1
"$work/trimmed-mean-demo" -data "$work/million.csv" -cpuprofile results/cpu.prof -memprofile results/mem.prof > /dev/null 2>&1
{ echo "# CPU profile: one run of -data <1,000,000 generated values> -trim 0.05"; go tool pprof -top -nodecount=15 "$work/trimmed-mean-demo" results/cpu.prof 2>/dev/null | sed '/^File:/d;/^Build ID:/d;/^Type:/d;/^Time:/d'; } > results/cpu-profile.txt
{ echo "# Memory profile (allocated during the run): same run"; go tool pprof -top -sample_index=alloc_space -nodecount=15 "$work/trimmed-mean-demo" results/mem.prof 2>/dev/null | sed '/^File:/d;/^Build ID:/d;/^Type:/d;/^Time:/d'; } > results/mem-profile.txt
{ echo "# Largest committed sample ($largest): too short for the CPU sampler (100 Hz); kept to show it"; go tool pprof -top -nodecount=5 "$work/trimmed-mean-demo" results/cpu-largest-sample.prof 2>/dev/null | sed '/^File:/d;/^Build ID:/d;/^Time:/d'; } >> results/cpu-profile.txt

echo "measure: step 4 of 5: machine"
{
  echo "measured: $(date '+%Y-%m-%d %H:%M %Z')"
  echo "commit: $(git rev-parse --short HEAD)"
  echo "machine: $(sysctl -n hw.model) · $(sysctl -n machdep.cpu.brand_string) · $(( $(sysctl -n hw.memsize) / 1073741824 )) GB · macOS $(sw_vers -productVersion)"
  echo "go: $(go version)"
  echo "trimmedmean: $package_version"
  echo "R reference: $(head -n 1 testdata/r-reference/r-session.txt | sed 's/^R: //'), $(sed -n 2p testdata/r-reference/r-session.txt | sed 's/^platform: //')"
} > results/machine.txt
cat results/machine.txt

echo "measure: step 5 of 5: tables"
scripts/readme.sh
echo "measure: done. Commit and push results/ and README.md."

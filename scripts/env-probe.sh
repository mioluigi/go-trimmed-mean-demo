#!/usr/bin/env bash
# env-probe.sh: SFEDD §5.7 environment probe, Mac side.
#
# Records what the official machine has (Go, git, R, Go module settings, module proxy) in
# docs/spikes/env-mac.txt. It changes nothing outside that file and a temporary folder, installs
# nothing, and records no user name, host name or home path.
set -uo pipefail
cd "$(dirname "$0")/.."
out=docs/spikes/env-mac.txt
fail() { echo "env-probe: $*" >&2; exit 1; }
[ "$(uname -s)" = "Darwin" ] || fail "this probe is for the Mac (this is $(uname -s))."

# have NAME: prints the first line of "NAME --version", or MISSING.
have() {
  if command -v "$1" >/dev/null 2>&1; then "$@" 2>&1 | head -n 1; else echo "MISSING"; fi
}

{
  echo "probe: $(date '+%Y-%m-%d %H:%M %Z')"
  echo "machine: $(sysctl -n hw.model) · $(sysctl -n machdep.cpu.brand_string) · $(( $(sysctl -n hw.memsize) / 1073741824 )) GB · macOS $(sw_vers -productVersion)"
  echo "cpus: $(sysctl -n hw.ncpu)"
  echo "go: $(have go version)"
  if command -v go >/dev/null 2>&1; then
    echo "go env: GOPROXY=$(go env GOPROXY) GOSUMDB=$(go env GOSUMDB) GOPRIVATE=$(go env GOPRIVATE) GOFLAGS=$(go env GOFLAGS)"
  fi
  echo "git: $(have git --version)"
  echo "gh (GitHub CLI): $(have gh --version)"
  echo "R: $(have R --version)"
  echo "Rscript: $(command -v Rscript >/dev/null 2>&1 && echo present || echo MISSING)"
  if command -v Rscript >/dev/null 2>&1; then
    echo "R sizeof long double: $(Rscript -e 'cat(.Machine$sizeof.longdouble)' 2>/dev/null)"
    echo "R mean(c(1, 2, 3, 4, 100), trim = 0.2): $(Rscript -e 'cat(mean(c(1, 2, 3, 4, 100), trim = 0.2))' 2>/dev/null)"
  fi
  echo "/usr/bin/time -l: $([ -x /usr/bin/time ] && echo present || echo MISSING)"

  scratch="$(mktemp -d)"
  if command -v go >/dev/null 2>&1 \
    && (cd "$scratch" && go mod init envprobe >/dev/null 2>&1 \
        && go get github.com/seehuhn/mt19937@v1.0.0 >/dev/null 2>&1); then
    echo "go module proxy (go get github.com/seehuhn/mt19937@v1.0.0): ok"
  else
    echo "go module proxy (go get github.com/seehuhn/mt19937@v1.0.0): FAILED"
  fi
  # Spike 001 P8: can the Mac fetch one of Luis's own public modules through the proxy?
  if command -v go >/dev/null 2>&1 \
    && own_module="$(cd "$scratch" && go list -m github.com/mioluigi/go-applied-statistics@latest 2>&1)"; then
    echo "own module through proxy (go list -m …/go-applied-statistics@latest): ok · $own_module"
  else
    echo "own module through proxy (go list -m …/go-applied-statistics@latest): FAILED"
  fi
  rm -rf "$scratch"
} > "$out"

cat "$out"
echo
echo "env-probe: wrote $out"

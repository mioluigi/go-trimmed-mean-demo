#!/usr/bin/env bash
# check-fresh.sh: the measurement freshness guard (SFEDD §5.3, L-W6-03). Succeeds only when this copy has
# no uncommitted changes and is exactly what GitHub has, so results always name a commit that exists.
set -euo pipefail
cd "$(dirname "$0")/.."
fail() { echo "check-fresh: $*" >&2; exit 1; }
[ -z "$(git status --porcelain)" ] || fail "uncommitted changes; commit or stash them, then run again."
git fetch -q origin
[ "$(git rev-parse HEAD)" = "$(git rev-parse '@{u}')" ] || fail "this copy differs from GitHub; run git pull (or push), then run again."
echo "check-fresh: ok ($(git rev-parse --short HEAD) matches GitHub)"

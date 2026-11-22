#!/usr/bin/env bash
# readme.sh: regenerates results/tables.md and the README's generated blocks from results/ (cmd/tables).
# With --check it changes nothing and fails if the README is out of date.
set -euo pipefail
cd "$(dirname "$0")/.."
if [ "${1:-}" = "--check" ]; then
  go run ./cmd/tables -check
else
  go run ./cmd/tables
fi

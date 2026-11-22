#!/usr/bin/env bash
# build.sh: executables for macOS (Apple silicon and Intel) and Windows (spec NFR-5).
#
# The program and the trimmedmean package are pure Go with no other dependencies, so they cross-compile
# from any machine. Output folder: dist/ (or the folder in DIST). The macOS .app is built by
# scripts/build-macos-app.sh. Run a binary from the repository folder, where data/ and testdata/ are.
set -euo pipefail
cd "$(dirname "$0")/.."
dist="${DIST:-dist}"
mkdir -p "$dist"
CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -trimpath -o "$dist/trimmed-mean-demo-macos-arm64" .
CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -trimpath -o "$dist/trimmed-mean-demo-macos-amd64" .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o "$dist/trimmed-mean-demo.exe" .
ls -l "$dist"

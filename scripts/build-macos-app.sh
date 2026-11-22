#!/usr/bin/env bash
# build-macos-app.sh: TrimmedMean.app, a double-clickable macOS application (adapted from Weeks 6–8).
# Output folder: dist/ (or the folder in DIST).
#
# The program is a command-line tool and a double-clicked app has no terminal, so the bundle is a thin
# wrapper: its launcher opens Terminal, runs the comparison with R on the samples bundled inside the
# app, writes the full table to the user's temporary folder, shows how to analyse another file, and
# waits for Return so the result stays on screen. The program itself is unchanged.
set -euo pipefail
cd "$(dirname "$0")/.."
dist="${DIST:-dist}"
app="$dist/TrimmedMean.app"
resources="$app/Contents/Resources"
rm -rf "$app"
mkdir -p "$app/Contents/MacOS" "$resources/data" "$resources/testdata/r-reference"
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -o "$resources/trimmed-mean-demo" .
cp data/*.csv "$resources/data/"
cp testdata/r-reference/trimmed-means.tsv "$resources/testdata/r-reference/"

cat >"$resources/run-in-terminal.command" <<'SCRIPT'
#!/bin/bash
clear
cd "$(dirname "$0")" || exit 1
table="${TMPDIR:-/tmp}/trimmed-mean-comparison.tsv"
echo "Trimmed means in Go (github.com/mioluigi/trimmedmean) compared with R."
echo
if ./trimmed-mean-demo -out "$table" 2>/dev/null; then
  echo
  echo "To compute a trimmed mean of your own file (one number per line), run in Terminal:"
  echo "  \"$PWD/trimmed-mean-demo\" -data /path/to/your-file.csv -trim 0.05"
  echo "  (or -trim 0.05,0.10 to trim 5% from the bottom and 10% from the top; -h lists every option)"
else
  echo
  echo "The program reported a problem; see the message above."
fi
echo
read -r -p "Press Return to close this window. " _
SCRIPT

cat >"$app/Contents/MacOS/TrimmedMean" <<'SCRIPT'
#!/bin/bash
open -a Terminal "$(dirname "$0")/../Resources/run-in-terminal.command"
SCRIPT
chmod +x "$app/Contents/MacOS/TrimmedMean" "$resources/run-in-terminal.command"

cat >"$app/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>               <string>TrimmedMean</string>
  <key>CFBundleDisplayName</key>        <string>Trimmed Mean</string>
  <key>CFBundleIdentifier</key>         <string>com.github.mioluigi.trimmedmean-demo</string>
  <key>CFBundleExecutable</key>         <string>TrimmedMean</string>
  <key>CFBundlePackageType</key>        <string>APPL</string>
  <key>CFBundleShortVersionString</key> <string>1.0</string>
  <key>CFBundleVersion</key>            <string>1</string>
  <key>LSMinimumSystemVersion</key>     <string>13.0</string>
  <key>LSArchitecturePriority</key>     <array><string>arm64</string></array>
</dict>
</plist>
PLIST
echo "Built $app. Double-click it in Finder, or run: open $app"

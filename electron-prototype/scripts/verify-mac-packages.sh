#!/bin/bash
set -euo pipefail
if [ "$(uname -s)" != Darwin ]; then echo 'Run this verification on macOS.' >&2; exit 1; fi
export GITHUB_SHA="${GITHUB_SHA:-$(git rev-parse HEAD)}"
for directory in release-build release-build-macos12; do
  minimum=13.0
  runtime=44.4.5
  if [ "$directory" = release-build-macos12 ]; then minimum=12.0; runtime=43.7.5; fi
  count=0
  while IFS= read -r -d '' app; do
    expected_arch=x86_64
    case "$app" in */mac-arm64/*) expected_arch=arm64 ;; esac
    actual_arch="$(lipo -archs "$app/Contents/MacOS/MofuMouse")"
    actual_minimum="$(/usr/libexec/PlistBuddy -c 'Print :LSMinimumSystemVersion' "$app/Contents/Info.plist")"
    echo "Verifying $app: architecture=$actual_arch minimum=$actual_minimum"
    test "$actual_arch" = "$expected_arch"
    test "$actual_minimum" = "$minimum"
    codesign --verify --deep --strict "$app"
    count=$((count + 1))
  done < <(find "$directory" -maxdepth 2 -name '*.app' -print0)
  test "$count" -eq 2
  appdir=mac
  if [ "$(node -p process.arch)" = arm64 ]; then appdir=mac-arm64; fi
  ELECTRON_RUN_AS_NODE=1 "$directory/$appdir/MofuMouse.app/Contents/MacOS/MofuMouse" -e "if(process.versions.electron!=='$runtime')process.exit(1)"
done
node scripts/verify-release-build.mjs
node scripts/verify-release-build.mjs --monterey

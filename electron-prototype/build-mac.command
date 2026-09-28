#!/bin/bash
set -euo pipefail
cd -- "$(dirname -- "$0")"
if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This command builds on macOS. Use npm run pack:win on Windows." >&2
  exit 1
fi
if ! command -v node >/dev/null || ! command -v npm >/dev/null; then
  echo "Install Node.js 22.12 or newer, then run this command again." >&2
  exit 1
fi
node -e 'const [a,b]=process.versions.node.split(".").map(Number);if(a<22||(a===22&&b<12))throw Error("Node.js 22.12 or newer is required")'
unset ELECTRON_RUN_AS_NODE
npm ci --no-audit --no-fund
npm run check
arch="$(node -p 'process.arch')"
node scripts/package.mjs darwin "$arch"
app="release/MofuMouseElectron-darwin-$arch/MofuMouseElectron.app"
node scripts/smoke.mjs "$app/Contents/MacOS/MofuMouseElectron"
open "$app" --args --count=10 --size=48
echo "Started $app. Use the animal icon in the menu bar to change size/count or quit."

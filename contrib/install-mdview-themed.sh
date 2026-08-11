#!/usr/bin/env bash
# Installs mdview-themed: copies the script to ~/.local/bin and downloads
# marked + mermaid into ~/.local/share/mdview-themed for offline rendering.
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
BIN="$HOME/.local/bin"
LIB="$HOME/.local/share/mdview-themed"

mkdir -p "$BIN" "$LIB"
install -m 755 "$HERE/mdview-themed" "$BIN/mdview-themed"

[ -f "$LIB/marked.min.js" ]  || curl -fsSL -o "$LIB/marked.min.js"  "https://cdn.jsdelivr.net/npm/marked/marked.min.js"
[ -f "$LIB/mermaid.min.js" ] || curl -fsSL -o "$LIB/mermaid.min.js" "https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"

echo "installed: $BIN/mdview-themed"
echo "try:       mdview-themed README.md"
echo "theme:     mdview-themed --install-theme mdview-theme.json"

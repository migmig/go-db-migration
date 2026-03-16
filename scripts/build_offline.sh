#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
EMBED_DIR="$ROOT_DIR/internal/web/assets/v16"
PLACEHOLDER_INDEX="$EMBED_DIR/index.html"
OUTPUT_PATH="${1:-$ROOT_DIR/dbmigrator}"
NPM_BIN="${NPM_BIN:-npm}"
GO_BIN="${GO_BIN:-go}"

if [[ ! -f "$PLACEHOLDER_INDEX" ]]; then
  echo "missing placeholder asset: $PLACEHOLDER_INDEX" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$EMBED_DIR"
  mkdir -p "$EMBED_DIR"
  cp "$TMP_DIR/index.html" "$PLACEHOLDER_INDEX"
  rm -rf "$FRONTEND_DIR/dist"
  rm -f "$FRONTEND_DIR"/tsconfig*.tsbuildinfo
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

cp "$PLACEHOLDER_INDEX" "$TMP_DIR/index.html"

if [[ ! -d "$FRONTEND_DIR/node_modules" ]]; then
  echo "[offline] installing frontend dependencies"
  (cd "$FRONTEND_DIR" && "$NPM_BIN" ci)
fi

echo "[offline] verifying frontend"
(cd "$FRONTEND_DIR" && "$NPM_BIN" run verify:fast)

echo "[offline] staging embedded frontend"
rm -rf "$EMBED_DIR"
mkdir -p "$EMBED_DIR"
cp -R "$FRONTEND_DIR/dist/." "$EMBED_DIR/"

echo "[offline] building binary -> $OUTPUT_PATH"
(cd "$ROOT_DIR" && "$GO_BIN" build -o "$OUTPUT_PATH" .)

echo "[offline] done"

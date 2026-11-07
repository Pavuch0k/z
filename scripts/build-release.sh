#!/usr/bin/env bash
# Сборка транслятора zc для Linux и Windows.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${VERSION:-0.1.0}"
OUT="$ROOT/bin"
mkdir -p "$OUT"

echo "build-release: version ${VERSION}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags "-X main.Version=${VERSION}" \
  -o "$OUT/zc-${VERSION}-linux-amd64" \
  "$ROOT/cmd/zc"

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags "-X main.Version=${VERSION}" \
  -o "$OUT/zc-${VERSION}-windows-amd64.exe" \
  "$ROOT/cmd/zc"

echo "build-release: ok"
ls -la "$OUT/zc-${VERSION}-linux-amd64" "$OUT/zc-${VERSION}-windows-amd64.exe"
"$OUT/zc-${VERSION}-linux-amd64" version

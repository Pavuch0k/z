#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
echo "=== go test ./... ==="
go test ./...
echo "=== examples via zc ==="
go build -o /tmp/zc-test ./cmd/zc
/tmp/zc-test run examples/hello.z >/tmp/zc-hello.out
diff -u <(printf 'Hello, Z\n') /tmp/zc-hello.out
/tmp/zc-test run examples/bad_string.z >/dev/null 2>&1 && exit 1 || true
/tmp/zc-test run examples/bad_mul_string.z >/dev/null 2>&1 && exit 1 || true
echo "OK"

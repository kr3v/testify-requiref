#!/usr/bin/env bash
# Fails if require/generated.go or require/assertions_generated.go is stale with
# respect to the testify version in go.mod.
set -euo pipefail
cd "$(dirname "$0")"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

cp require/generated.go require/assertions_generated.go "$tmp/"
go run ./internal/gen require

for f in generated.go assertions_generated.go; do
  if ! diff -u "$tmp/$f" "require/$f"; then
    echo "require/$f is stale -- run: go generate ./require/" >&2
    exit 1
  fi
done
echo "codegen up to date"

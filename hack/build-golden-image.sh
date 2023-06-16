#!/usr/bin/env bash

set -euo pipefail

NAME="${1:-}"
GOLDBLD_DIR="./hack/goldbld"

if [[ -z "${NAME}" ]]; then
  echo "Usage: $0 <name>"
  exit 1
fi

mkdir -p "$GOLDBLD_DIR"
go build -o "$GOLDBLD_DIR/$NAME" "./cmd/agib"

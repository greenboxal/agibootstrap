#!/bin/bash

set -euo pipefail

DIST_DIR="$(realpath ./dist)"

rm -rf "$DIST_DIR"

echo "Building..."
env BABEL_ENV=development npx webpack

for name in "$DIST_DIR"/*; do
  echo "Deploying $name"

  MOD_NAME="$(basename "$name" .bundle.js)"

  http --verify=no \
    POST "https://0.0.0.0:22440/v1/psi/$MOD_NAME" \
    Accept:application/json \
    type==vm.Module \
    "name=$MOD_NAME" "source=@$name"
done

#!/bin/bash

set -euo pipefail

. "$PSIDB_SDK/sdk/psidb-bash/psidb.sh"

DIST_DIR="$(realpath .)/dist"

rm -rf "$DIST_DIR" > /dev/null 2>&1 || true

echo "Building..."
env BABEL_ENV=development npx webpack

for name in "$DIST_DIR"/*; do
  echo "Deploying $name"

  MOD_NAME="$(basename "$name" .bundle.js)"

  psidb_create_node "QmYXZ//$MOD_NAME" type==psidb.vm.Module name="$MOD_NAME" source="@$name"
done

psidb_rpc_node QmYXZ//main IModule Test

#!/usr/bin/env bash

set -euo pipefail

bkpdir="./.fti/bkp/$(date +%Y%m%d%H%M%S)"

mkdir -p "$bkpdir"

echo "$bkpdir"
mv .fti/index "$bkpdir/"
mv .fti/objects "$bkpdir/"
mv .fti/index.faiss "$bkpdir/"
cp .fti/config.json "$bkpdir/"

go run ./cmd/fti update
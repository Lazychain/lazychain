#!/bin/bash

# There are some tool requirements here like "just", "cargo make"

set -o errexit -o nounset -o pipefail

echo "Get the repos for the contracts we need"

if [ ! -d "cw-ics721" ]; then
  git clone https://github.com/public-awesome/cw-ics721.git
fi

if [ ! -d "cw-nfts" ]; then
  git clone https://github.com/CosmWasm/cw-nfts.git
fi

# TODO: Fix so the architecture is not hardcoded
cd cw-ics721
git checkout v0.1.9
just optimize
cp artifacts/ics721_base-aarch64.wasm ../artifacts/ics721_base.wasm
cd ..

cd cw-nfts
cargo make optimize
cp artifacts/cw721_base-aarch64.wasm ../artifacts/cw721_base.wasm
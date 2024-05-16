#!/bin/bash
set -o errexit -o nounset -o pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

sleep 1
echo "## Submit a CosmWasm gov proposal"
RESP=$(slothd tx wasm submit-proposal store-instantiate "$DIR/../../x/wasm/keeper/testdata/reflect.wasm" \
  '{}' --label="testing" \
  --title "testing" --summary "Testing" --deposit "1000000000ustake" \
  --admin $(slothd keys show -a validator --keyring-backend=test) \
  --amount 123ustake \
  --keyring-backend=test \
  --from validator --gas auto --gas-adjustment=1.5 -y  --chain-id=testing --node=http://localhost:26657 -b sync -o json)
echo $RESP
sleep 6
slothd q tx $(echo "$RESP"| jq -r '.txhash') -o json | jq


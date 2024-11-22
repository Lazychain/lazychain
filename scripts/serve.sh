#!/bin/bash

set -e

# set variables for the chain
VALIDATOR_NAME=validator1

echo -e "\n Deleting existing $BINARY data... \n"
rm -rf ~/.lazychain/

DA_BLOCK_HEIGHT=$(curl public-celestia-mocha4-consensus.numia.xyz:26657/block |jq -r '.result.block.header.height')
AUTH_TOKEN=$(celestia light auth write --p2p.network mocha)

echo -e "\n Your DA_BLOCK_HEIGHT is $DA_BLOCK_HEIGHT \n"
echo -e "\n Your DA AUTH_TOKEN is $AUTH_TOKEN \n"

rollkit init gg

ADDRESS=$(jq -r '.address' ~/.lazychain/config/priv_validator_key.json)
PUB_KEY=$(jq -r '.pub_key' ~/.lazychain/config/priv_validator_key.json)
jq --argjson pubKey "$PUB_KEY" '.consensus["validators"]=[{"address": "'$ADDRESS'", "pub_key": $pubKey, "power": "1", "name": "Rollkit Sequencer"}]' ~/.lazychain/config/genesis.json > temp.json && mv temp.json ~/.lazychain/config/genesis.json
PUB_KEY_VALUE=$(jq -r '.pub_key .value' ~/.lazychain/config/priv_validator_key.json)
jq --arg pubKey $PUB_KEY_VALUE '.app_state .sequencer["sequencers"]=[{"name": "test-1", "consensus_pubkey": {"@type": "/cosmos.crypto.ed25519.PubKey","key":$pubKey}}]' ~/.lazychain/config/genesis.json >temp.json && mv temp.json ~/.lazychain/config/genesis.json

# create a restart-testnet.sh file to restart the chain later
[ -f restart-lazychain.sh ] && rm restart-lazychain.sh
echo "DA_BLOCK_HEIGHT=$DA_BLOCK_HEIGHT" >> restart-lazychain.sh
echo "AUTH_TOKEN=$AUTH_TOKEN" >> restart-lazychain.sh

echo "rollkit start --rollkit.lazy_aggregator --rollkit.aggregator --rollkit.da_auth_token=\$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height \$DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr \"0.0.0.0:26656\" --minimum-gas-prices="0stake"  --api.enable --api.enabled-unsafe-cors" >> restart-lazychain.sh
chmod +x restart-lazychain.sh

rollkit genesis validate
rollkit start --rollkit.aggregator --rollkit.da_auth_token=$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height $DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr "0.0.0.0:26656" --minimum-gas-prices="0stake" --api.enable --api.enabled-unsafe-cors
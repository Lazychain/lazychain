#!/bin/sh

set -e

# set variables for the chain
VALIDATOR_NAME=validator1
CHAIN_ID=sloth
BINARY=lazychaind
KEY_NAME=slothy
TOKEN_AMOUNT=10000000000000000000000000stake
STAKING_AMOUNT=1000000000stake
RELAYER_ADDRESS=lazy1avl4q6s02pss5q2ftrkjqaft3jk75q4ldesnwe

echo -e "\n Deleting existing $BINARY data... \n"
rm -rf ~/.lazychain/

echo -e "\n Installing the chain...\n"
make install

# query the DA Layer start height, in this case we are querying
# an RPC endpoint provided by Celestia Labs. The RPC endpoint is
# to allow users to interact with Celestia's core network by querying
# the node's state and broadcasting transactions on the Celestia
# network.

# Mocha
DA_BLOCK_HEIGHT=$(curl public-celestia-mocha4-consensus.numia.xyz:26657/block |jq -r '.result.block.header.height')
AUTH_TOKEN=$(celestia light auth write --p2p.network mocha)
# Arabica
#DA_BLOCK_HEIGHT=$(curl https://rpc.celestia-arabica-11.com/block |jq -r '.result.block.header.height')
#AUTH_TOKEN=$(celestia light auth write --p2p.network arabica)

echo -e "\n Your DA_BLOCK_HEIGHT is $DA_BLOCK_HEIGHT \n"
echo -e "\n Your DA AUTH_TOKEN is $AUTH_TOKEN \n"

# reset any existing genesis/chain data
$BINARY tendermint unsafe-reset-all
$BINARY init $VALIDATOR_NAME --chain-id $CHAIN_ID

# update $BINARY configuration files to set chain details and enable necessary settings
# the sed commands here are editing various configuration settings for the $BINARY instance
# such as setting minimum gas prices, enabling the api, setting the chain id, setting the rpc address,
# adjusting time constants, and setting the denomination for bonds and minting.
sed -i'' -e 's/^minimum-gas-prices *= .*/minimum-gas-prices = "0stake"/' "$HOME"/.lazychain/config/app.toml
sed -i'' -e '/\[api\]/,+3 s/enable *= .*/enable = true/' "$HOME"/.lazychain/config/app.toml
sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$HOME"/.lazychain/config/client.toml
sed -i'' -e '/\[rpc\]/,+3 s/laddr *= .*/laddr = "tcp:\/\/0.0.0.0:26657"/' "$HOME"/.lazychain/config/config.toml
sed -i'' -e 's/"time_iota_ms": "1000"/"time_iota_ms": "10"/' "$HOME"/.lazychain/config/genesis.json
sed -i'' -e 's/bond_denom": ".*"/bond_denom": "stake"/' "$HOME"/.lazychain/config/genesis.json
sed -i'' -e 's/mint_denom": ".*"/mint_denom": "stake"/' "$HOME"/.lazychain/config/genesis.json

# add a key to keyring-backend test
$BINARY keys add $KEY_NAME --keyring-backend test

# add a genesis account
$BINARY genesis add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test
$BINARY genesis add-genesis-account $RELAYER_ADDRESS $TOKEN_AMOUNT

# set the staking amounts in the genesis transaction
$BINARY genesis gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

# collect gentxs
$BINARY genesis collect-gentxs

# copy centralized sequencer address into genesis.json
# Note: validator and sequencer are used interchangeably here
ADDRESS=$(jq -r '.address' ~/.lazychain/config/priv_validator_key.json)
PUB_KEY=$(jq -r '.pub_key' ~/.lazychain/config/priv_validator_key.json)
jq --argjson pubKey "$PUB_KEY" '.consensus["validators"]=[{"address": "'$ADDRESS'", "pub_key": $pubKey, "power": "1", "name": "Rollkit Sequencer"}]' ~/.lazychain/config/genesis.json > temp.json && mv temp.json ~/.lazychain/config/genesis.json
PUB_KEY_VALUE=$(jq -r '.pub_key .value' ~/.lazychain/config/priv_validator_key.json)
jq --arg pubKey $PUB_KEY_VALUE '.app_state .sequencer["sequencers"]=[{"name": "test-1", "consensus_pubkey": {"@type": "/cosmos.crypto.ed25519.PubKey","key":$pubKey}}]' ~/.lazychain/config/genesis.json >temp.json && mv temp.json ~/.lazychain/config/genesis.json


# create a restart-testnet.sh file to restart the chain later
[ -f restart-$BINARY.sh ] && rm restart-$BINARY.sh
echo "DA_BLOCK_HEIGHT=$DA_BLOCK_HEIGHT" >> restart-$BINARY.sh
echo "AUTH_TOKEN=$AUTH_TOKEN" >> restart-$BINARY.sh

echo "$BINARY start --rollkit.lazy_aggregator --rollkit.aggregator --rollkit.da_auth_token=\$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height \$DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr \"0.0.0.0:26656\" --minimum-gas-prices="0stake"  --api.enable --api.enabled-unsafe-cors" >> restart-$BINARY.sh

# start the chain
$BINARY genesis validate
$BINARY start --rollkit.aggregator --rollkit.da_auth_token=$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height $DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr "0.0.0.0:26656" --minimum-gas-prices="0stake" --api.enable --api.enabled-unsafe-cors
#!/bin/sh

# set variables for the chain
VALIDATOR_NAME=validator1
CHAIN_ID=sloth
KEY_NAME=slothy
TOKEN_AMOUNT="10000000000000000000000000ulazy"
STAKING_AMOUNT=1000000000ulazy
CHAINFLAG="--chain-id ${CHAIN_ID}"
TXFLAG="--chain-id ${CHAIN_ID} --gas-prices 0ulazy --gas auto --gas-adjustment 1.3"

echo -e "\n Deleting existing slothd data... \n"
rm -rf $HOME"/.slothd"

# query the DA Layer start height, in this case we are querying
# an RPC endpoint provided by Celestia Labs. The RPC endpoint is
# to allow users to interact with Celestia's core network by querying
# the node's state and broadcasting transactions on the Celestia
# network. This is for Arabica, if using another network, change the RPC.
DA_BLOCK_HEIGHT=$(curl https://rpc-mocha.pops.one/block |jq -r '.result.block.header.height')
echo -e "\n Your DA_BLOCK_HEIGHT is $DA_BLOCK_HEIGHT \n"

AUTH_TOKEN=$(celestia light auth write --p2p.network mocha)
echo -e "\n Your DA AUTH_TOKEN is $AUTH_TOKEN \n"

# reset any existing genesis/chain data
slothd tendermint unsafe-reset-all
slothd init $VALIDATOR_NAME --chain-id $CHAIN_ID

# update slothd configuration files to set chain details and enable necessary settings
# the sed commands here are editing various configuration settings for the slothd instance
# such as setting minimum gas prices, enabling the api, setting the chain id, setting the rpc address,
# adjusting time constants, and setting the denomination for bonds and minting.
sed -i'' -e 's/^minimum-gas-prices *= .*/minimum-gas-prices = "0ulazy"/' "$HOME"/.slothd/config/app.toml
sed -i'' -e '/\[api\]/,+3 s/enable *= .*/enable = true/' "$HOME"/.slothd/config/app.toml
sed -i'' -e "s/^chain-id *= .*/chain-id = \"$CHAIN_ID\"/" "$HOME"/.slothd/config/client.toml
sed -i'' -e '/\[rpc\]/,+3 s/laddr *= .*/laddr = "tcp:\/\/0.0.0.0:26657"/' "$HOME"/.slothd/config/config.toml
sed -i'' -e 's/"time_iota_ms": "1000"/"time_iota_ms": "10"/' "$HOME"/.slothd/config/genesis.json
sed -i'' -e 's/bond_denom": ".*"/bond_denom": "ulazy"/' "$HOME"/.slothd/config/genesis.json
sed -i'' -e 's/mint_denom": ".*"/mint_denom": "ulazy"/' "$HOME"/.slothd/config/genesis.json

# add a key to keyring-backend test
slothd keys add $KEY_NAME --keyring-backend test

# add a genesis account
slothd genesis add-genesis-account $KEY_NAME $TOKEN_AMOUNT --keyring-backend test

# set the staking amounts in the genesis transaction
slothd genesis gentx $KEY_NAME $STAKING_AMOUNT --chain-id $CHAIN_ID --keyring-backend test

# collect gentxs
slothd genesis collect-gentxs

# copy centralized sequencer address into genesis.json
# Note: validator and sequencer are used interchangeably here
ADDRESS=$(jq -r '.address' ~/.slothd/config/priv_validator_key.json)
PUB_KEY=$(jq -r '.pub_key' ~/.slothd/config/priv_validator_key.json)
jq --argjson pubKey "$PUB_KEY" '.consensus["validators"]=[{"address": "'$ADDRESS'", "pub_key": $pubKey, "power": "1000", "name": "Rollkit Sequencer"}]' ~/.slothd/config/genesis.json > temp.json && mv temp.json ~/.slothd/config/genesis.json

# create a restart-testnet.sh file to restart the chain later
[ -f restart-slothd.sh ] && rm restart-slothd.sh
echo "DA_BLOCK_HEIGHT=$DA_BLOCK_HEIGHT" >> restart-slothd.sh
echo "AUTH_TOKEN=$AUTH_TOKEN" >> restart-slothd.sh

echo "slothd start --rollkit.lazy_aggregator --rollkit.aggregator --rollkit.da_auth_token=\$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height \$DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr \"0.0.0.0:26656\" --minimum-gas-prices="0.025ulazy"  --api.enable --api.enabled-unsafe-cors" >> restart-slothd.sh

# start the chain
slothd start --rollkit.lazy_aggregator --rollkit.aggregator --rollkit.da_auth_token=$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height $DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr "0.0.0.0:26656" --minimum-gas-prices="0.025ulazy" --api.enable --api.enabled-unsafe-cors

#!/bin/zsh

# We expect the chain the be running at this point

set -o errexit -o nounset -o pipefail

echo "Get the repos for DAO DAO stuff"

if [ ! -d "dao-dao-ui" ]; then
  git clone git@github.com:DA0-DA0/dao-dao-ui.git
fi

if [ ! -d "dao-contracts" ]; then
  git clone git@github.com:DA0-DA0/dao-contracts.git
fi

if [ ! -d "cw-plus" ]; then
  git clone https://github.com/CosmWasm/cw-plus
fi

declare -A code_id_map

cd cw-plus
echo "If you need to build the CW Plus contracts, there are some commands here that needs to be uncommented"
#git pull
#docker run --rm -v "$(pwd)":/code \
#  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
#  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
#  cosmwasm/workspace-optimizer:0.13.0

for file in ./artifacts/*.wasm
do
  STORE_TX_HASH=$(slothd tx wasm store "$file" --from slothy --gas-prices 0.025ulazy --gas auto --gas-adjustment 1.75 --chain-id sloth --yes --keyring-backend test --output json | jq -r ".txhash")
  sleep 3
  CODE_ID=$(slothd q tx "$STORE_TX_HASH" --output json | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
  code_id_map[$file]=$CODE_ID

  echo "Uploaded $file with code id $CODE_ID"
done

cd ..

cd dao-contracts
echo "If you need to build the DAO DAO contracts, there are some commands here that needs to be uncommented"
#git pull
#echo "Build the DAO DAO contracts"
#just workspace-optimize

for file in ./artifacts/*.wasm
do
  STORE_TX_HASH=$(slothd tx wasm store "$file" --from slothy --gas-prices 0.025ulazy --gas auto --gas-adjustment 1.75 --chain-id sloth --yes --keyring-backend test --output json | jq -r ".txhash")
  sleep 3
  CODE_ID=$(slothd q tx "$STORE_TX_HASH" --output json | jq -r '.events[] | select(.type=="store_code") | .attributes[] | select(.key=="code_id") | .value')
  code_id_map[$file]=$CODE_ID

  echo "Uploaded $file with code id $CODE_ID"
done

FACTORY_CODE_ID=${code_id_map[./artifacts/cw_admin_factory-aarch64.wasm]}
FACTORY_INSTANTIATE_TX_HASH=$(slothd tx wasm instantiate "$FACTORY_CODE_ID" '{}' --label cw_admin_factory --no-admin --from slothy --gas-prices 0.025ulazy --gas auto --gas-adjustment 1.75 --chain-id sloth --yes --keyring-backend test --output json | jq -r ".txhash")
sleep 3
CONTRACT_ADDRESS=$(slothd q tx "$FACTORY_INSTANTIATE_TX_HASH" --output json | jq -r '.events[] | select(.type=="instantiate") | .attributes[] | select(.key=="_contract_address") | .value')
echo "Factory contract address: $CONTRACT_ADDRESS"

cat <<EOF
  {
    chainId: ChainId.SlothChainDevnet,
    name: 'slothchain',
    mainnet: false,
    accentColor: '#00d9ff',
    factoryContractAddress:
        '$CONTRACT_ADDRESS',
    explorerUrlTemplates: {
      tx: 'https://testnet.ping.pub/stargaze/tx/REPLACE',
      gov: 'https://testnet.ping.pub/stargaze/gov',
      govProp: 'https://testnet.ping.pub/stargaze/gov/REPLACE',
      wallet: 'https://testnet.ping.pub/stargaze/account/REPLACE',
    },
    codeIds: {
      // https://github.com/CosmWasm/cw-plus
      Cw1Whitelist: ${code_id_map[./artifacts/cw1_whitelist.wasm]},
      Cw4Group: ${code_id_map[./artifacts/cw4_group.wasm]},

      CwPayrollFactory: ${code_id_map[./artifacts/cw_payroll_factory-aarch64.wasm]},
      CwTokenSwap: ${code_id_map[./artifacts/cw_token_swap-aarch64.wasm]},
      CwTokenfactoryIssuerMain: ${code_id_map[./artifacts/cw_tokenfactory_issuer-aarch64.wasm]},
      CwVesting: ${code_id_map[./artifacts/cw_vesting-aarch64.wasm]},
      DaoCore: ${code_id_map[./artifacts/dao_dao_core-aarch64.wasm]},
      DaoMigrator: -0, // not needed since only v2 DAOs exist
      DaoPreProposeApprovalSingle: ${code_id_map[./artifacts/dao_pre_propose_approval_single-aarch64.wasm]},
      DaoPreProposeApprover: ${code_id_map[./artifacts/dao_pre_propose_approver-aarch64.wasm]},
      DaoPreProposeMultiple: ${code_id_map[./artifacts/dao_pre_propose_multiple-aarch64.wasm]},
      DaoPreProposeSingle: ${code_id_map[./artifacts/dao_pre_propose_single-aarch64.wasm]},
      DaoProposalMultiple: ${code_id_map[./artifacts/dao_proposal_multiple-aarch64.wasm]},
      DaoProposalSingle: ${code_id_map[./artifacts/dao_proposal_single-aarch64.wasm]},
      DaoVotingCw4: ${code_id_map[./artifacts/dao_voting_cw4-aarch64.wasm]},
      DaoVotingCw721Staked: ${code_id_map[./artifacts/dao_voting_cw721_staked-aarch64.wasm]},
      DaoVotingTokenStaked: ${code_id_map[./artifacts/dao_voting_token_staked-aarch64.wasm]},
    },
    historicalCodeIds: {
      [ContractVersion.V210]: {
        DaoPreProposeMultiple: 224,
        DaoProposalMultiple: 226,
      },
    },
  }
EOF

#cd ..

#cd dao-dao-ui
#git pull

#echo "Set up the UI"
#yarn
#yarn build
#!/bin/zsh

declare -A code_id_map

CONTRACT_ADDRESS="ctrcaddr"

code_id_map[./artifacts/cw1_whitelist.wasm]=1
code_id_map[./artifacts/cw4_group.wasm]=2

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
EOF
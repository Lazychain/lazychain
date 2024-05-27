DA_BLOCK_HEIGHT=1110638
AUTH_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJwdWJsaWMiLCJyZWFkIiwid3JpdGUiXX0.V3qt6dm5qD3Sdzi9yO-YIbKdZ0ar5I-OEmqZ-mphZ6c
slothchaind start --rollkit.lazy_aggregator --rollkit.aggregator --rollkit.da_auth_token=$AUTH_TOKEN --rollkit.da_namespace 00000000000000000000000000000000000000000008e5f679bf7116cb --rollkit.da_start_height $DA_BLOCK_HEIGHT --rpc.laddr tcp://127.0.0.1:26657 --grpc.address 127.0.0.1:9290 --p2p.laddr "0.0.0.0:26656" --minimum-gas-prices=0.025ulazy  --api.enable --api.enabled-unsafe-cors

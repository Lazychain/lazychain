# SlothChain 🦥

LM... 🦥💤

## Installation

```bash
$ ignite chain build
```

## Interslothtest

The interslothtest directory contains an e2e test suite for the Slothchain IBC setup:
- Slothchain
- Stargaze
- Celestia

The test suite uses Interchaintest to spin up a full environment with ICS721 and all you need to test the full
sloth journey end-to-end.

You can run the test suite with the following command:
```bash
$ cd interslothtest
$ go test -v -p 1 ./...
```

### Run a lazy 💤 local interslothchain environment

The repo has a very lazy option if you want to run a full local environment with a single command.

The environment consists of:
- Slothchain (duh... 🦥)
- Stargaze
- Celestia
- Relayer

The environment sets up all the above components and configures:
- User with funds on all chains (mnemonic: `curve govern feature draw giggle one enemy shop wonder cross castle oxygen business obscure rule detail chaos dirt pause parrot tail lunch merit rely`)
- An NFT contract on Stargaze (to mimic the Sloth collection)
- ICS721 deployed on Stargaze and Slothchain
- IBC connection between Slothchain and Stargaze
- Channels for both ICS20 between all chains
- Channels for ICS721 between Slothchain and Stargaze

To transfer, see the command section below.

There are some pre-requisites to run the interslothchain environment:
- Go
- Docker
- slothchain:local image built (`make local-docker`)

To run it:
```bash
$ cd interslothtest
$ go run ./localinterslothchain
```

It takes a while to spin up everything, deploy the contracts and whatnot, but once it is finished it will output something like following:
```shell
Users, all with the mnemonic: curve govern feature draw giggle one enemy shop wonder cross castle oxygen business obscure rule detail chaos dirt pause parrot tail lunch merit rely
Sloth user address: lazy1ct9r7k20kp7z2m90066h6h2anq0rvmmrhwcl0w
Stargaze user address: stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk
Celestia user address: celestia1ct9r7k20kp7z2m90066h6h2anq0rvmmrtnldz2

Slothchain chain-id: slothtestchain-1
Slothchain RPC address: tcp://localhost:63921
Stargaze chain-id: stargazetest-1
Stargaze RPC address: tcp://localhost:63910
Celestia chain-id: celestiatest-1
Celestia RPC address: tcp://localhost:63915

ICS721 setup deployed
ICS721 contract on Stargaze: stars1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrq096cja
ICS721 contract on Sloth chain: lazy1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtq8xhtac
Sloth contract: stars14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9srsl6sm
Stargaze to Sloth channel: channel-1
Sloth chain to Stargaze channel: channel-2
Celestia to Sloth channel: channel-0

Press Ctrl+C to stop...
```

### Commands

Slothchaind has some built-in commands to interact directly with the sloths.
- `slothchaind tx sloths transfer [from] [to] [nft-id] [--flags]`
- `slothchaind q sloths owned-by [address] [--flags]`

Both commands have a `--mainnet` and `--testnet` flag to fill in all the necessary flags for the respective chain.
They are not implemented at the moment, but will be once testnet and mainnet are live.

#### Transfer transaction

This command will transfer an NFT from one chain to another using ICS721 (it supports both Stargaze->Slothchain and Slothchain->Stargaze).
It does not currently support transfer between two addresses on the same chain.

With `--mainnet` or `--testnet` flag:
```bash
$ slothchaind tx sloths transfer stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk 1 --testnet
```

With all override flags (necessary for local interslothchain):
```bash
$ slothchaind tx sloths transfer stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk lazy1ct9r7k20kp7z2m90066h6h2anq0rvmmrhwcl0w 1 --node tcp://localhost:57023 --chain-id stargazetest-1 --nft-contract stars14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9srsl6sm --ics721-contract stars1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrq096cja --ics721-channel channel-1 --gas auto --gas-adjustment 1.5 --keyring-backend test
```

#### Owned By query

This command will query all the NFTs owned by a specific address on a specific chain.

With `--mainnet` or `--testnet` flag:
```bash
$ slothchaind q sloths owned-by stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk --mainnet
```

With all override flags (necessary for local interslothchain):
```bash
$ slothchaind q sloths owned-by stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk --node tcp://localhost:57023 --nft-contract stars14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9srsl6sm
```
## 💤
Too... Lazy... To... Write... More... 🦥
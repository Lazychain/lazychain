# LazyChain 🦥

LM... 🦥💤

## Installation

```bash
$ make install
```

## Interchaintest

The interchaintest directory contains an e2e test suite for the LazyChain IBC setup:
- LazyChain
- Stargaze
- Celestia

The test suite uses Interchaintest to spin up a full environment with ICS721 and all you need to test the full
sloth journey end-to-end.

First you need to build the local docker image:
```bash
$ make local-docker
```

> You can also specify the image with environment variables `LAZYCHAIN_IMAGE_REPOSITORY` and `LAZYCHAIN_IMAGE_VERSION`.
> For instance, you can run the latest built in CI with `LAZYCHAIN_IMAGE_REPOSITORY=ghcr.io/Lazychain/lazychain` and `LAZYCHAIN_IMAGE_VERSION=latest`.

You can run the test suite with the following command:
```bash
$ cd interchaintest
$ go test -v -p 1 ./...
```

### Run a lazy 💤 local interlazychain environment

The repo has a very lazy option if you want to run a full local environment with a single command.

The environment consists of:
- LazyChain (duh... 🦥)
- Stargaze
- Celestia
- Relayer

The environment sets up all the above components and configures:
- User with funds on all chains (mnemonic: `curve govern feature draw giggle one enemy shop wonder cross castle oxygen business obscure rule detail chaos dirt pause parrot tail lunch merit rely`)
- An NFT contract on Stargaze (to mimic the Sloth collection)
- ICS721 deployed on Stargaze and LazyChain
- IBC connection between LazyChain and Stargaze
- Channels for both ICS20 between all chains
- Channels for ICS721 between LazyChain and Stargaze

To transfer, see the command section below.

There are some pre-requisites to run the interlazychain environment:
- Go
- Docker
- lazychain:local image built (`make local-docker`)

To run it:
```bash
$ cd interchaintest
$ go run ./local-interchain
```

It takes a while to spin up everything, deploy the contracts and whatnot, but once it is finished it will output something like following:
```shell
Users, all with the mnemonic: curve govern feature draw giggle one enemy shop wonder cross castle oxygen business obscure rule detail chaos dirt pause parrot tail lunch merit rely
Sloth user address: lazy1ct9r7k20kp7z2m90066h6h2anq0rvmmrhwcl0w
Stargaze user address: stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk
Celestia user address: celestia1ct9r7k20kp7z2m90066h6h2anq0rvmmrtnldz2

LazyChain chain-id: lazytestchain-1
LazyChain RPC address: tcp://localhost:63921
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

The `lazychaind` binary has some built-in commands to interact directly with the sloths.
- `lazychaind tx sloths transfer [from] [to] [nft-id] [--flags]`
- `lazychaind q sloths owned-by [address] [--flags]`

Both commands have a `--mainnet` and `--testnet` flag to fill in all the necessary flags for the respective chain.
They are not implemented at the moment, but will be once testnet and mainnet are live.

#### Transfer transaction

This command will transfer an NFT from one chain to another using ICS721 (it supports both Stargaze->LazyChain and LazyChain->Stargaze).
It does not currently support transfer between two addresses on the same chain.

With `--mainnet` or `--testnet` flag from Stargaze to LazyChain
```bash
$ lazychaind tx sloths transfer stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk lazy1u0g894r00fu3rnh7ft35yzk9smyaxscyhax3vs 1 --testnet
```

With `--mainnet` or `--testnet` flag from LazyChain to Stargaze
```bash
$ lazychaind tx sloths transfer lazy1u0g894r00fu3rnh7ft35yzk9smyaxscyhax3vs stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk 1 --testnet 
```

With all override flags (necessary for local interchain environment)
```bash
$ lazychaind tx sloths transfer stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk lazy1ct9r7k20kp7z2m90066h6h2anq0rvmmrhwcl0w 1 --node tcp://localhost:57023 --chain-id stargazetest-1 --nft-contract stars14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9srsl6sm --ics721-contract stars1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrq096cja --ics721-channel channel-1 --gas auto --gas-adjustment 1.5 --keyring-backend test
```

#### Owned By query

This command will query all the NFTs owned by a specific address on a specific chain.

With `--mainnet` or `--testnet` flag:
```bash
$ lazychaind q sloths owned-by stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk --testnet
```

With all override flags (necessary for local interlazychain):
```bash
$ lazychaind q sloths owned-by stars1ct9r7k20kp7z2m90066h6h2anq0rvmmrw9eqnk --node tcp://localhost:57023 --nft-contract stars14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9srsl6sm
```
## 💤
Too... Lazy... To... Write... More... 🦥

package sloths

const (
	flagWaitForTx = "wait-for-tx"
	flagMainnet   = "mainnet"
	flagTestnet   = "testnet"

	// If none of the above are provided, we can use the flags below
	flagNFTContract    = "nft-contract"
	flagICS721Contract = "ics721-contract"
	flagICS721Channel  = "ics721-channel"
)

type Networks struct {
	Slotchain StaticNetworkInfo
	Stargaze  StaticNetworkInfo
}

type StaticNetworkInfo struct {
	ChainID        string
	Node           string
	NFTContract    string
	ICS721Contract string
	ICS721Channel  string
}

var (
	Testnet = Networks{
		Slotchain: StaticNetworkInfo{
			ChainID:        "lazynet-1",
			Node:           "tcp://51.159.101.58:26657",
			NFTContract:    "lazy167pjcglw3pusa9kheavpc4ujnpzc0w7jfue092nssd7hq2ku43cq8fqc8c",
			ICS721Contract: "lazy1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqnzqkf5",
			ICS721Channel:  "channel-2",
		},
		Stargaze: StaticNetworkInfo{
			ChainID:        "elgafar-1",
			Node:           "https://rpc.elgafar-1.stargaze-apis.com:443",
			NFTContract:    "stars1egctj79q59t68pvcwfuz3fhy3mncs95z7gk4dpmh2t4w5rc8h27q5zn2eg", // TODO: Update with actual contract address
			ICS721Contract: "stars1n2nejlcr3758rh5yfsg8jkq7xv6kv9ls00wu5ja66qcw26ka4npszpng5v", // TODO: Update to use SG proxy, this is base
			ICS721Channel: "channel-980",
		},
	}
)

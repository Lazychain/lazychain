package lazycommandutils

type ICS721Networks struct {
	LazyChain StaticICS721NetworkInfo
	Stargaze  StaticICS721NetworkInfo
}

type StaticICS721NetworkInfo struct {
	ChainID        string
	Node           string
	GasPrices      string
	NFTContract    string
	ICS721Contract string
	ICS721Channel  string
}

var (
	ICS721Mainnets = ICS721Networks{} // TODO: Update once mainnet
	ICS721Testnets = ICS721Networks{
		LazyChain: StaticICS721NetworkInfo{
			ChainID:        testnetLazyChainChainID,
			Node:           testnetLazyChainNode,
			GasPrices:      testnetLazyChainGasPrices,
			NFTContract:    "lazy1enc8mxs5tsu6zzmvu0uh9snj28yp4nt2xycv3l6054p339gweavqgd224r",
			ICS721Contract: "lazy1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtq8xhtac",
			ICS721Channel:  "channel-1",
			// ICS721Contract: "lazy1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqnzqkf5", // Base, connected with channel-1 and Base ICS721 contract
			// ICS721Channel:  "channel-2", // Connected with base version of ICS721
		},
		Stargaze: StaticICS721NetworkInfo{
			ChainID:        testnetStargazeChainID,
			Node:           testnetStargazeNode,
			GasPrices:      testnetStargazeGasPrices,
			NFTContract:    "stars1z5vs00kvjwr3h050twnul5pd2sftk42ajy7xp6sg59gux4cvprvq08eryl",
			ICS721Contract: "stars1338rc4fn2r3k9z9x783pmtgcwcqmz5phaksurrznnu9dnu4dmctqr2gyzl", // Proxy for SG-721 (wasm.stars1cxnwk637xwee9gcw0v2ua00gnyhvzxkte8ucnxzfxj0ea8nxkppsgacht3)
			ICS721Channel:  "channel-979",                                                      // SG
			// ICS721Contract: "stars1n2nejlcr3758rh5yfsg8jkq7xv6kv9ls00wu5ja66qcw26ka4npszpng5v", // Base ICS721 contract
			// ICS721Channel:  "channel-980", // Base ICS721 channel
		},
	}
)

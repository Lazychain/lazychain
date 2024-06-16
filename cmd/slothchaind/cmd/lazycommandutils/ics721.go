package lazycommandutils

type Networks struct {
	Slotchain StaticICS721NetworkInfo
	Stargaze  StaticICS721NetworkInfo
}

type StaticICS721NetworkInfo struct {
	ChainID        string
	Node           string
	NFTContract    string
	ICS721Contract string
	ICS721Channel  string
}

var ICS721Testnets = Networks{
	Slotchain: StaticICS721NetworkInfo{
		ChainID:        testnetSlothchainChainID,
		Node:           testnetSlothchainNode,
		NFTContract:    "lazy167pjcglw3pusa9kheavpc4ujnpzc0w7jfue092nssd7hq2ku43cq8fqc8c",
		ICS721Contract: "lazy1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrqnzqkf5",
		ICS721Channel:  "channel-2",
	},
	Stargaze: StaticICS721NetworkInfo{
		ChainID:        testnetStargazeChainID,
		Node:           testnetStargazeNode,
		NFTContract:    "stars1egctj79q59t68pvcwfuz3fhy3mncs95z7gk4dpmh2t4w5rc8h27q5zn2eg", // TODO: Update with actual contract address
		ICS721Contract: "stars1n2nejlcr3758rh5yfsg8jkq7xv6kv9ls00wu5ja66qcw26ka4npszpng5v", // TODO: Update to use SG proxy, this is base
		ICS721Channel:  "channel-980",
	},
}

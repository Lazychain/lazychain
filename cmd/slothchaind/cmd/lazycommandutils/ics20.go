package lazycommandutils

type ICS20Networks struct {
	Celestia   StaticICS20NetworkInfo
	Slothchain StaticICS20NetworkInfo
}

type StaticICS20NetworkInfo struct {
	ChainID      string
	Node         string
	ICS20Denom   string
	ICS20Channel string
}

var ICS20Testnets = ICS20Networks{
	Celestia: StaticICS20NetworkInfo{
		ChainID:      testnetCelestiaChainID,
		Node:         testnetCelestiaNode,
		ICS20Denom:   "utia",
		ICS20Channel: "channel-95",
	},
	Slothchain: StaticICS20NetworkInfo{
		ChainID:      testnetSlothchainChainID,
		Node:         testnetSlothchainNode,
		ICS20Denom:   "ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
		ICS20Channel: "channel-0",
	},
}

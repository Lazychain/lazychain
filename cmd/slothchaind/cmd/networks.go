package cmd

type Networks struct {
	Slotchain StaticNetworkInfo
	Stargaze StaticNetworkInfo
	Celestia StaticNetworkInfo
}

type StaticNetworkInfo struct {
	ChainID string
	NFTContract string
	ICS721Contract string
	ICS721Channel string
}

var (
	Testnet = Networks{
		Slotchain: StaticNetworkInfo{
			ChainID:        "lazynet-1",
			NFTContract:    "",
			ICS721Contract: "",
			ICS721Channel:  "",
		},
		Stargaze: StaticNetworkInfo{
			ChainID:        "elgafar-1",
			NFTContract:    "",
			ICS721Contract: "",
			ICS721Channel:  "",
		},
		Celestia: StaticNetworkInfo{
			ChainID:        "mocha-4",
			NFTContract:    "",
			ICS721Contract: "",
			ICS721Channel:  "",
		},
	}
)

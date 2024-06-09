package utils

import (
	"context"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
	"strings"
)

// Not const because we need to give them as pointers later
var (
	slothVals          = 1
	slothFullNodes     = 0
	celestialVals      = 1
	celestialFullNodes = 0
	stargazeVals       = 1
	stargazeFullNodes  = 0

	votingPeriod     = "15s"
	maxDepositPeriod = "10s"

	slothChainId    = "slothtestchain-1"
	celestiaChainID = "celestiatest-1"
	sgChainID       = "stargazetest-1"
)

type E2ETestSuite struct {
	testifysuite.Suite

	Ctx               context.Context
	Interchain        *interchaintest.Interchain
	Network           string
	Relayer           ibc.Relayer
	RelayerExecRep    *testreporter.RelayerExecReporter
	CelestiaSlothPath string
	StargazeSlothPath string

	Slothchain *cosmos.CosmosChain
	Celestia   *cosmos.CosmosChain
	Stargaze   *cosmos.CosmosChain
}

func (s *E2ETestSuite) SetupSuite() {
	s.Ctx = context.Background()

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	s.Interchain = ic
	cf := s.getChainFactory()
	chains, err := cf.Chains(s.T().Name())
	slothchain, celestia, stargaze := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)
	s.NoError(err)
	s.Slothchain = slothchain
	s.Celestia = celestia
	s.Stargaze = stargaze

	for _, chain := range chains {
		ic.AddChain(chain)
	}

	client, network := interchaintest.DockerSetup(s.T())

	rf := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(s.T()),
		interchaintestrelayer.CustomDockerImage("ghcr.io/cosmos/relayer", "latest", "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	r := rf.Build(s.T(), client, network)
	s.Relayer = r

	ic.AddRelayer(r, "relayer")
	s.StargazeSlothPath = "sg-sloth-path"
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  stargaze,
		Chain2:  slothchain,
		Relayer: r,
		Path:    s.StargazeSlothPath,
	})
	s.CelestiaSlothPath = "celestia-sloth-path"
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  celestia,
		Chain2:  slothchain,
		Relayer: r,
		Path:    s.CelestiaSlothPath,
	})

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(s.T())
	s.RelayerExecRep = eRep

	err = ic.Build(s.Ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         s.T().Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	})
	s.NoError(err)

	// For some reason automated path creation in Build didn't work when doing two paths 🤷
	s.NoError(s.Relayer.GeneratePath(s.Ctx, s.RelayerExecRep, s.Celestia.Config().ChainID, s.Slothchain.Config().ChainID, s.CelestiaSlothPath))
	s.NoError(s.Relayer.LinkPath(s.Ctx, s.RelayerExecRep, s.CelestiaSlothPath, ibc.DefaultChannelOpts(), ibc.DefaultClientOpts()))

	s.NoError(s.Relayer.GeneratePath(s.Ctx, s.RelayerExecRep, s.Stargaze.Config().ChainID, s.Slothchain.Config().ChainID, s.StargazeSlothPath))
	s.NoError(s.Relayer.LinkPath(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath, ibc.DefaultChannelOpts(), ibc.DefaultClientOpts()))

	s.T().Cleanup(func() {
		_ = ic.Close()
	})
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	if s.Interchain != nil {
		_ = s.Interchain.Close()
	}
}

func (s *E2ETestSuite) getChainFactory() *interchaintest.BuiltinChainFactory {
	return interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(s.T()), []*interchaintest.ChainSpec{
		{
			Name:      "slothchain",
			ChainName: "slothchain",
			Version:   "local",
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "slothchain",
				ChainID: slothChainId,
				Images: []ibc.DockerImage{
					{
						Repository: "slothchain",
						Version:    "local",
						UidGid:     "1025:1025",
					},
				},
				Bin:                 "slothchaind",
				Bech32Prefix:        "lazy",
				Denom:               "useq",
				CoinType:            "118",
				GasPrices:           "0ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
				GasAdjustment:       2.0,
				TrustingPeriod:      "112h",
				NoHostMount:         false,
				ConfigFileOverrides: nil,
				EncodingConfig:      getEncodingConfig(),
				ModifyGenesisAmounts: func(_ int) (sdk.Coin, sdk.Coin) {
					return sdk.NewInt64Coin("useq", 10_000_000_000_000), sdk.NewInt64Coin("useq", 1_000_000_000)
				},
				ModifyGenesis: func(config ibc.ChainConfig, bytes []byte) ([]byte, error) {
					addressBz, _, err := s.Slothchain.Validators[0].Exec(s.Ctx, []string{"jq", "-r", ".address", "/var/cosmos-chain/slothchain/config/priv_validator_key.json"}, []string{})
					if err != nil {
						return nil, err
					}
					address := strings.TrimSuffix(string(addressBz), "\n")
					pubKeyBz, _, err := s.Slothchain.Validators[0].Exec(s.Ctx, []string{"jq", "-r", ".pub_key.value", "/var/cosmos-chain/slothchain/config/priv_validator_key.json"}, []string{})
					if err != nil {
						return nil, err
					}
					pubKey := strings.TrimSuffix(string(pubKeyBz), "\n")

					newGenesis := []cosmos.GenesisKV{
						{
							Key: "consensus.validators",
							Value: []map[string]interface{}{
								{
									"address": address,
									"pub_key": map[string]interface{}{
										"type":  "tendermint/PubKeyEd25519",
										"value": pubKey,
									},
									"power": "1000",
									"name":  "Rollkit Sequencer",
								},
							},
						},
					}

					name := s.Slothchain.Sidecars[0].HostName()
					_, _, err = s.Slothchain.Validators[0].Exec(s.Ctx, []string{"sh", "-c", fmt.Sprintf(`echo "[rollkit]
da_address = \"http://%s:%s\"" >> /var/cosmos-chain/slothchain/config/config.toml`, name, "7980")}, []string{})
					if err != nil {
						return nil, err
					}

					return cosmos.ModifyGenesis(newGenesis)(config, bytes)
				},
				AdditionalStartArgs: []string{"--rollkit.aggregator", "true", "--api.enable", "--api.enabled-unsafe-cors", "--rpc.laddr", "tcp://0.0.0.0:26657"},
				SidecarConfigs: []ibc.SidecarConfig{
					{
						ProcessName: "mock-da",
						Image: ibc.DockerImage{
							Repository: "ghcr.io/gjermundgaraba/mock-da",
							Version:    "pessimist",
							UidGid:     "1025:1025",
						},
						HomeDir:          "",
						Ports:            []string{"7980/tcp"},
						StartCmd:         []string{"/usr/bin/mock-da", "-listen-all"},
						Env:              nil,
						PreStart:         true,
						ValidatorProcess: false,
					},
				},
			},
			NumValidators: &slothVals,
			NumFullNodes:  &slothFullNodes,
		},
		{
			Name:      "celestia",
			ChainName: "celestia",
			Version:   "v1.9.0",
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "celestia",
				ChainID: celestiaChainID,
				Images: []ibc.DockerImage{
					{
						Repository: "ghcr.io/strangelove-ventures/heighliner/celestia",
						Version:    "v1.9.0",
						UidGid:     "1025:1025",
					},
				},
				Bin:            "celestia-appd",
				Bech32Prefix:   "celestia",
				Denom:          "utia",
				CoinType:       "118",
				GasPrices:      "0utia",
				GasAdjustment:  2.0,
				TrustingPeriod: "112h",
				NoHostMount:    false,
				ConfigFileOverrides: map[string]any{
					"config/config.toml": testutil.Toml{
						"storage": testutil.Toml{
							"discard_abci_responses": false,
						},
						"tx_index": testutil.Toml{
							"indexer": "kv",
						},
					},
					"config/app.toml": testutil.Toml{
						"grpc": testutil.Toml{
							"enable": true,
						},
					},
				},
				EncodingConfig: getEncodingConfig(),
			},
			NumValidators: &celestialVals,
			NumFullNodes:  &celestialFullNodes,
		},
		{
			Name:      "stargaze",
			ChainName: "stargaze",
			Version:   "v13.0.0",
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "stargaze",
				ChainID:        sgChainID,
				CoinType:       "118",
				GasPrices:      "0stars",
				GasAdjustment:  2.0,
				EncodingConfig: getEncodingConfig(),
			},
			NumValidators: &stargazeVals,
			NumFullNodes:  &stargazeFullNodes,
		},
	})
}

func getEncodingConfig() *sdktestutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	// whatever.RegisterInterfaces(cfg.InterfaceRegistry)
	wasm.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

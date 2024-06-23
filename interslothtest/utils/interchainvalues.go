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
	"os"
	"strings"
)

type InterchainValues struct {
	// We hold a reference to this, so we can implement the s.NoError and get access to the T() method
	testifySuiteRef *testifysuite.Suite
	fakeT           *FakeT

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

func (s *InterchainValues) SetupFakeT(name string) {
	if name == "" {
		name = "fake_test_name"
	}

	s.fakeT = &FakeT{
		FakeName: name,
	}
}

func (s *InterchainValues) GetFakeT() *FakeT {
	return s.fakeT
}

func (s *InterchainValues) TT() CustomT {
	if s.testifySuiteRef != nil {
		return s.testifySuiteRef.T()
	}

	return s.fakeT
}

func (s *InterchainValues) NoError(err error) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.NoError(err)
		return
	}

	if err != nil {
		panic(err)
	}
}

func (s *InterchainValues) ErrorContains(err error, contains string) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.ErrorContains(err, contains)
		return
	}

	if err == nil {
		panic("error is nil")
	}

	if !strings.Contains(err.Error(), contains) {
		panic(fmt.Sprintf("error does not contain %s", contains))
	}
}

func (s *InterchainValues) NotEmpty(value interface{}) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.NotEmpty(value)
		return
	}

	if value == nil {
		panic("value is empty")
	}
}

func (s *InterchainValues) Len(value interface{}, length int) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.Len(value, length)
		return
	}

	// TODO: Check better plz
	if value == nil {
		panic("value is empty")
	}
}

func (s *InterchainValues) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.Equal(expected, actual, msgAndArgs)
		return
	}

	if expected != actual {
		panic(fmt.Sprintf("expected: %v, actual: %v", expected, actual))
	}
}

func (s *InterchainValues) SetupInterchainValues() {
	s.Ctx = context.Background()

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	s.Interchain = ic
	cf := s.getChainFactory()
	chains, err := cf.Chains(s.TT().Name())
	slothchain, celestia, stargaze := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)
	s.NoError(err)
	s.Slothchain = slothchain
	s.Celestia = celestia
	s.Stargaze = stargaze

	for _, chain := range chains {
		ic.AddChain(chain)
	}

	client, network := interchaintest.DockerSetup(s.TT())

	rf := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(s.TT()),
		interchaintestrelayer.CustomDockerImage("ghcr.io/cosmos/relayer", "latest", "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	r := rf.Build(s.TT(), client, network)
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
	eRep := rep.RelayerExecReporter(s.TT())
	s.RelayerExecRep = eRep

	err = ic.Build(s.Ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         s.TT().Name(),
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

	s.TT().Cleanup(func() {
		_ = ic.Close()
	})
}

func (s *InterchainValues) getChainFactory() *interchaintest.BuiltinChainFactory {
	slothchainImageRepository := "slothchain"
	slothchainImageVersion := "local"
	envImageVersion, found := os.LookupEnv("SLOTHCHAIN_IMAGE_VERSION")
	if found {
		s.TT().Log("SLOTHCHAIN_IMAGE_VERSION from environment found", envImageVersion)
		slothchainImageVersion = envImageVersion
	}
	envImageRepository, found := os.LookupEnv("SLOTHCHAIN_IMAGE_REPOSITORY")
	if found {
		s.TT().Log("SLOTHCHAIN_IMAGE_REPOSITORY from environment found", envImageRepository)
		slothchainImageRepository = envImageRepository
	}

	s.TT().Log("SLOTHCHAIN_IMAGE_VERSION", slothchainImageVersion)
	s.TT().Log("SLOTHCHAIN_IMAGE_REPOSITORY", slothchainImageRepository)

	return interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(s.TT()), []*interchaintest.ChainSpec{
		{
			Name:      "slothchain",
			ChainName: "slothchain",
			Version:   slothchainImageVersion,
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "slothchain",
				ChainID: slothChainId,
				Images: []ibc.DockerImage{
					{
						Repository: slothchainImageRepository,
						Version:    slothchainImageVersion,
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

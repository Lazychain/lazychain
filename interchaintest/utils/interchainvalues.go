package utils

import (
	"context"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/pelletier/go-toml/v2"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/hermes"
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

	LazyChain *cosmos.CosmosChain
	Celestia  *cosmos.CosmosChain
	Stargaze  *cosmos.CosmosChain
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

func (s *InterchainValues) True(value bool) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.True(value)
		return
	}

	if !value {
		panic("value is not true")
	}
}

func (s *InterchainValues) NotNil(value interface{}) {
	if s.testifySuiteRef != nil {
		s.testifySuiteRef.NotNil(value)
		return
	}

	if value == nil {
		panic("value is nil")
	}
}

func (s *InterchainValues) SetupInterchainValues() {
	s.Ctx = context.Background()

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	s.Interchain = ic
	cf := s.getChainFactory()
	chains, err := cf.Chains(s.TT().Name())
	lazyChain, celestia, stargaze := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain)
	s.NoError(err)
	s.LazyChain = lazyChain
	s.Celestia = celestia
	s.Stargaze = stargaze

	for _, chain := range chains {
		ic.AddChain(chain)
	}

	client, network := interchaintest.DockerSetup(s.TT())

	rf := interchaintest.NewBuiltinRelayerFactory(
		//ibc.Hermes,
		ibc.CosmosRly,
		zaptest.NewLogger(s.TT()),
		//interchaintestrelayer.CustomDockerImage("ghcr.io/informalsystems/hermes", "1.10.0", "2000:2000"),
		interchaintestrelayer.CustomDockerImage("ghcr.io/cosmos/relayer", "latest", "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	)
	r := rf.Build(s.TT(), client, network)
	s.Relayer = r

	ic.AddRelayer(r, "relayer")
	s.StargazeSlothPath = "sg-sloth-path"
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  stargaze,
		Chain2:  lazyChain,
		Relayer: r,
		Path:    s.StargazeSlothPath,
	})
	s.CelestiaSlothPath = "celestia-sloth-path"
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  celestia,
		Chain2:  lazyChain,
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

	//s.modifyHermesConfig(r.(*hermes.Relayer))
	//res := s.Relayer.Exec(s.Ctx, eRep, []string{"cat", "/home/hermes/.hermes/config.toml"}, nil)
	//s.TT().Log(string(res.Stdout))

	// For some reason automated path creation in Build didn't work when doing two paths 🤷
	s.NoError(s.Relayer.GeneratePath(s.Ctx, s.RelayerExecRep, s.Celestia.Config().ChainID, s.LazyChain.Config().ChainID, s.CelestiaSlothPath))
	s.NoError(s.Relayer.LinkPath(s.Ctx, s.RelayerExecRep, s.CelestiaSlothPath, ibc.DefaultChannelOpts(), ibc.DefaultClientOpts()))

	s.NoError(s.Relayer.GeneratePath(s.Ctx, s.RelayerExecRep, s.Stargaze.Config().ChainID, s.LazyChain.Config().ChainID, s.StargazeSlothPath))
	s.NoError(s.Relayer.LinkPath(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath, ibc.DefaultChannelOpts(), ibc.DefaultClientOpts()))

	s.TT().Cleanup(func() {
		_ = ic.Close()
	})
}

func (s *InterchainValues) getChainFactory() *interchaintest.BuiltinChainFactory {
	lazyChainImageRepository := "lazychain"
	lazyChainImageVersion := "local"
	envImageVersion, found := os.LookupEnv("LAZYCHAIN_IMAGE_VERSION")
	if found {
		s.TT().Log("LAZYCHAIN_IMAGE_VERSION from environment found", envImageVersion)
		lazyChainImageVersion = envImageVersion
	}
	envImageRepository, found := os.LookupEnv("LAZYCHAIN_IMAGE_REPOSITORY")
	if found {
		s.TT().Log("LAZYCHAIN_IMAGE_REPOSITORY from environment found", envImageRepository)
		lazyChainImageRepository = envImageRepository
	}

	s.TT().Log("LAZYCHAIN_IMAGE_VERSION", lazyChainImageVersion)
	s.TT().Log("LAZYCHAIN_IMAGE_REPOSITORY", lazyChainImageRepository)

	return interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(s.TT()), []*interchaintest.ChainSpec{
		{
			Name:      "lazychain",
			ChainName: "lazychain",
			Version:   lazyChainImageVersion,
			ChainConfig: ibc.ChainConfig{
				Type:    "cosmos",
				Name:    "lazychain",
				ChainID: lazyChainId,
				Images: []ibc.DockerImage{
					{
						Repository: lazyChainImageRepository,
						Version:    lazyChainImageVersion,
						UidGid:     "1025:1025",
					},
				},
				Bin:                 "lazychaind",
				Bech32Prefix:        "lazy",
				Denom:               "useq",
				CoinType:            "118",
				GasPrices:           "0.00useq",
				GasAdjustment:       2.0,
				TrustingPeriod:      "112h",
				NoHostMount:         false,
				ConfigFileOverrides: nil,
				EncodingConfig:      getEncodingConfig(),
				ModifyGenesisAmounts: func(_ int) (sdk.Coin, sdk.Coin) {
					return sdk.NewInt64Coin("useq", 10_000_000_000_000), sdk.NewInt64Coin("useq", 1_000_000_000)
				},
				ModifyGenesis: func(config ibc.ChainConfig, bytes []byte) ([]byte, error) {
					addressBz, _, err := s.LazyChain.Validators[0].Exec(s.Ctx, []string{"jq", "-r", ".address", "/var/cosmos-chain/lazychain/config/priv_validator_key.json"}, []string{})
					if err != nil {
						return nil, err
					}
					address := strings.TrimSuffix(string(addressBz), "\n")

					pubKeyBz, _, err := s.LazyChain.Validators[0].Exec(s.Ctx, []string{"jq", "-r", ".pub_key.value", "/var/cosmos-chain/lazychain/config/priv_validator_key.json"}, []string{})
					if err != nil {
						return nil, err
					}
					pubKey := strings.TrimSuffix(string(pubKeyBz), "\n")

					pubKeyValueBz, _, err := s.LazyChain.Validators[0].Exec(s.Ctx, []string{"jq", "-r", ".pub_key .value", "/var/cosmos-chain/lazychain/config/priv_validator_key.json"}, []string{})
					if err != nil {
						return nil, err
					}
					pubKeyValue := strings.TrimSuffix(string(pubKeyValueBz), "\n")

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
									"power": "1",
									"name":  "Rollkit Sequencer",
								},
							},
						},
						{
							Key: "app_state.sequencer.sequencers",
							Value: []map[string]interface{}{
								{
									"name": "test-1",
									"consensus_pubkey": map[string]interface{}{
										"@type": "/cosmos.crypto.ed25519.PubKey",
										"key":   pubKeyValue,
									},
								},
							},
						},
					}

					// '.app_state .sequencer["sequencers"]=[{"name": "test-1", "consensus_pubkey": {"@type": "/cosmos.crypto.ed25519.PubKey","key":$pubKey}}]'

					name := s.LazyChain.Sidecars[0].HostName()
					_, _, err = s.LazyChain.Validators[0].Exec(s.Ctx, []string{"sh", "-c", fmt.Sprintf(`echo "[rollkit]
da_address = \"http://%s:%s\"" >> /var/cosmos-chain/lazychain/config/config.toml`, name, "7980")}, []string{})
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
			NumValidators: &lazyVals,
			NumFullNodes:  &lazyFullNodes,
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
				GasPrices:      "0.00utia",
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
				GasPrices:      "0.00ustars",
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

func (s *InterchainValues) modifyHermesConfig(h *hermes.Relayer) {
	bz, err := h.ReadFileFromHomeDir(s.Ctx, ".hermes/config.toml")
	s.NoError(err)

	var config map[string]interface{}
	err = toml.Unmarshal(bz, &config)
	s.NoError(err)

	chains, ok := config["chains"].([]interface{})
	s.True(ok)
	var celestia, lazychain map[string]interface{}
	for _, ci := range chains {
		c, ok := ci.(map[string]interface{})
		s.True(ok)
		if c["id"] == celestiaChainID {
			celestia = c
		} else if c["id"] == lazyChainId {
			lazychain = c
		}
	}
	s.NotNil(celestia)

	celestia["compat_mode"] = "0.34"

	lazychain["event_source"] = map[string]interface{}{
		"mode":        "pull",
		"interval":    "1s",
		"max_retries": 20,
	}
	//lazychain["compat_mode"] = "0.37"

	bz, err = toml.Marshal(config)
	s.NoError(err)

	err = h.WriteFileToHomeDir(s.Ctx, ".hermes/config.toml", bz)
	s.NoError(err)
}

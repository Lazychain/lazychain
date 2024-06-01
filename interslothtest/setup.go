package interslothtest

import (
	"context"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchaintestrelayer "github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	testifysuite "github.com/stretchr/testify/suite"
	"go.uber.org/zap/zaptest"
	"strings"
)

// Not const because we need to give them as pointers later
var (
	slothVals         = 1
	slothFullNodes    = 0
	stargazeVals      = 1
	stargazeFullNodes = 0

	votingPeriod     = "15s"
	maxDepositPeriod = "10s"

	slothChainId = "slothtestchain-1"
	hubChainID   = "stargazetest-1"
)

type E2ETestSuite struct {
	testifysuite.Suite

	ctx         context.Context
	ic          *interchaintest.Interchain
	network     string
	r           ibc.Relayer
	eRep        *testreporter.RelayerExecReporter
	initialPath string

	slothchain *cosmos.CosmosChain
	stargaze   *cosmos.CosmosChain
}

func (s *E2ETestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain()
	s.ic = ic
	cf := s.getChainFactory()
	chains, err := cf.Chains(s.T().Name())
	slothchain, stargaze := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)
	s.NoError(err)
	s.slothchain = slothchain
	s.stargaze = stargaze

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
	s.r = r

	ic.AddRelayer(r, "relayer")
	s.initialPath = "ibc-path"
	ic.AddLink(interchaintest.InterchainLink{
		Chain1:  stargaze,
		Chain2:  slothchain,
		Relayer: r,
		Path:    s.initialPath,
	})

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(s.T())
	s.eRep = eRep

	err = ic.Build(s.ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         s.T().Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	})
	s.NoError(err)

	s.T().Cleanup(func() {
		_ = ic.Close()
	})
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	if s.ic != nil {
		_ = s.ic.Close()
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
				Denom:               "ulazy",
				CoinType:            "118",
				GasPrices:           "0ulazy",
				GasAdjustment:       2.0,
				TrustingPeriod:      "112h",
				NoHostMount:         false,
				ConfigFileOverrides: nil,
				EncodingConfig:      getEncodingConfig(),
				ModifyGenesisAmounts: func(_ int) (sdk.Coin, sdk.Coin) {
					return sdk.NewInt64Coin("ulazy", 10_000_000_000_000), sdk.NewInt64Coin("ulazy", 1_000_000_000)
				},
				ModifyGenesis: func(config ibc.ChainConfig, bytes []byte) ([]byte, error) {
					addressBz, _, err := s.slothchain.Validators[0].Exec(s.ctx, []string{"jq", "-r", ".address", "/var/cosmos-chain/slothchain/config/priv_validator_key.json"}, []string{})
					if err != nil {
						return nil, err
					}
					address := strings.TrimSuffix(string(addressBz), "\n")
					pubKeyBz, _, err := s.slothchain.Validators[0].Exec(s.ctx, []string{"jq", "-r", ".pub_key.value", "/var/cosmos-chain/slothchain/config/priv_validator_key.json"}, []string{})
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

					name := s.slothchain.Sidecars[0].HostName()
					_, _, err = s.slothchain.Validators[0].Exec(s.ctx, []string{"sh", "-c", fmt.Sprintf(`echo "[rollkit]
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
			Name:      "stargaze",
			ChainName: "stargaze",
			Version:   "v13.0.0",
			ChainConfig: ibc.ChainConfig{
				Type:                "cosmos",
				Name:                "stargaze",
				ChainID:             hubChainID,
				Bin:                 "starsd",
				Bech32Prefix:        "stars",
				Denom:               "stars",
				CoinType:            "118",
				GasPrices:           "0stars",
				GasAdjustment:       2.0,
				TrustingPeriod:      "112h",
				NoHostMount:         false,
				ConfigFileOverrides: nil,
				EncodingConfig:      getEncodingConfig(),
				ModifyGenesis: cosmos.ModifyGenesis([]cosmos.GenesisKV{
					{
						Key:   "app_state.gov.params.voting_period",
						Value: votingPeriod,
					},
					{
						Key:   "app_state.gov.params.max_deposit_period",
						Value: maxDepositPeriod,
					},
					{
						Key:   "app_state.gov.params.min_deposit.0.denom",
						Value: "stars",
					},
					{
						Key:   "app_state.gov.params.min_deposit.0.amount",
						Value: "1",
					},
				}),
			},
			NumValidators: &stargazeVals,
			NumFullNodes:  &stargazeFullNodes,
		},
	})
}

func getEncodingConfig() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	// whatever.RegisterInterfaces(cfg.InterfaceRegistry)
	wasm.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

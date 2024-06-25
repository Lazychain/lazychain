package ics721

import (
	"cosmossdk.io/math"
	"github.com/Lazychain/lazychain/interchaintest/utils"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"
	"testing"
)

type ICS721TestSuite struct {
	utils.E2ETestSuite
}

func TestICS721TestSuite(t *testing.T) {
	testifysuite.Run(t, new(ICS721TestSuite))
}

func (s *ICS721TestSuite) TestICS721() {
	users := interchaintest.GetAndFundTestUsers(s.T(), s.Ctx, s.T().Name(), math.NewInt(10_000_000_000), s.Stargaze, s.LazyChain)
	sgUser, slothUser := users[0], users[1]

	nftSetup := s.DeployNFTSetup(sgUser, slothUser, "../test-artifacts")

	s.NoError(s.Relayer.StartRelayer(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath))
	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.Stargaze, s.LazyChain))

	classID, lazyChainCW721 := s.TransferSlothToLazyChain(
		nftSetup,
		sgUser,
		slothUser,
		"1")
	_ = classID // wasm.lazy1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtq8xhtac/channel-2/stars14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9srsl6sm
	// wasm.lazyics721/slothchannel/starscontract

	s.AssertPacketRelayed(s.Stargaze, nftSetup.LazyChainPort, nftSetup.LazyChainChannel, 1)

	tokens := s.AllNFTs(s.LazyChain, lazyChainCW721)
	s.Len(tokens, 1)
	s.Equal("1", tokens[0]) // 🦥🚀

	s.TransferSlothToStargaze(nftSetup, slothUser, sgUser, "1", lazyChainCW721)
}

func (s *ICS721TestSuite) TestIncomingProxy() {
	users := interchaintest.GetAndFundTestUsers(s.T(), s.Ctx, s.T().Name(), math.NewInt(10_000_000_000), s.Stargaze, s.LazyChain)
	sgUser, slothUser := users[0], users[1]

	nftSetup := s.DeployNFTSetup(sgUser, slothUser, "../test-artifacts")
	nonSlothCW721CodeID := s.StoreCW721(s.Stargaze, sgUser.KeyName(), "../test-artifacts")
	nonSlothContractAddress := s.InstantiateCW721(nonSlothCW721CodeID, sgUser.KeyName(), "NOT A SLOTH", "NAS", sgUser.FormattedAddress())

	s.MintNFTs(nonSlothContractAddress, sgUser.KeyName(), sgUser.FormattedAddress(), []string{"1", "2", "3"})

	s.NoError(s.Relayer.StartRelayer(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath))
	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.Stargaze, s.LazyChain))

	err := s.TransferNFT(s.Stargaze, sgUser, slothUser, "1", nonSlothContractAddress, nftSetup.SGICS721Contract, nftSetup.SGChannel)
	s.NoError(err) // The transfer message itself on stargaze should succeed

	s.NoError(testutil.WaitForBlocks(s.Ctx, 10, s.Stargaze, s.LazyChain, s.Celestia))

	// Check that the token fails to actually transfer
	s.AssertPacketRelayed(s.Stargaze, nftSetup.LazyChainPort, nftSetup.LazyChainChannel, 1)

	cmd := "message.action='/ibc.core.channel.v1.MsgRecvPacket'"
	// cmd := "message.action=/ibc.core.channel.v1.MsgRecvPacket"
	txSearchRes, err := s.QueryTxsByEvents(s.LazyChain, 1, 10, cmd, "")
	s.Require().NoError(err)
	s.Require().Len(txSearchRes.Txs, 1)

	errorMessage, isFound := s.ExtractValueFromEvents(
		txSearchRes.Txs[0].Events,
		"write_acknowledgement",
		"packet_ack",
	)

	s.Require().True(isFound)
	s.Require().Equal(errorMessage, "{\"error\":\"codespace: wasm, code: 5\"}")

	type Response struct {
		Data [][]string `json:"data"`
	}
	var resp Response
	s.NoError(s.LazyChain.QueryContract(s.Ctx, nftSetup.LazyChainICS721Contract, "{\"nft_contracts\": {}}", &resp))
	s.Len(resp.Data, 0)
	// Update the incoming proxy with the non-sloth contract address to verify it works after the update
	s.MigrateICS721IncomingProxy(
		s.LazyChain,
		slothUser.KeyName(),
		"../test-artifacts",
		nftSetup.LazyChainICS721IncomingProxyContract,
		nonSlothContractAddress,
		nftSetup.LazyChainICS721Contract,
		nftSetup.LazyChainChannel,
	)

	err = s.TransferNFT(s.Stargaze, sgUser, slothUser, "1", nonSlothContractAddress, nftSetup.SGICS721Contract, nftSetup.SGChannel)
	s.NoError(err)

	s.NoError(testutil.WaitForBlocks(s.Ctx, 10, s.Stargaze, s.LazyChain, s.Celestia))

	// Check that the token fails to actually transfer
	s.AssertPacketRelayed(s.Stargaze, nftSetup.LazyChainPort, nftSetup.LazyChainChannel, 1)

	s.NoError(s.LazyChain.QueryContract(s.Ctx, nftSetup.LazyChainICS721Contract, "{\"nft_contracts\": {}}", &resp))
	s.Len(resp.Data, 1)

	tokens := s.AllNFTs(s.LazyChain, resp.Data[0][1])
	s.Len(tokens, 1)
	s.Equal("1", tokens[0]) // 🦥🚀
}

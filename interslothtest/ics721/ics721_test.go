package ics721

import (
	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"
	"interslothtest/utils"
	"testing"
)

type ICS721TestSuite struct {
	utils.E2ETestSuite
}

func TestICS721TestSuite(t *testing.T) {
	testifysuite.Run(t, new(ICS721TestSuite))
}

func (s *ICS721TestSuite) TestICS721() {
	users := interchaintest.GetAndFundTestUsers(s.T(), s.Ctx, s.T().Name(), math.NewInt(10_000_000_000), s.Stargaze, s.Slothchain)
	sgUser, slothUser := users[0], users[1]

	nftSetup := s.DeployNFTSetup(sgUser, slothUser, "../../artifacts")

	s.NoError(s.Relayer.StartRelayer(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath))
	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.Stargaze, s.Slothchain))

	classID, slothChainCW721 := s.TransferSlothToSlothChain(
		nftSetup,
		sgUser,
		slothUser,
		"1")
	_ = classID

	tokens := s.AllNFTs(slothChainCW721)
	s.Len(tokens, 1)
	s.Equal("1", tokens[0]) // 🦥🚀
}

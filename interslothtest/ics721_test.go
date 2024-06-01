package interslothtest

import (
	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"
	"testing"
)

type ICS721TestSuite struct {
	E2ETestSuite
}

func TestICS721TestSuite(t *testing.T) {
	testifysuite.Run(t, new(ICS721TestSuite))
}

func (s *ICS721TestSuite) TestICS721() {
	users := interchaintest.GetAndFundTestUsers(s.T(), s.ctx, s.T().Name(), math.NewInt(10_000_000_000), s.stargaze, s.slothchain)
	sgUser, slothUser := users[0], users[1]

	nftSetup := s.DeployNFTSetup(sgUser, slothUser)

	s.NoError(s.r.StartRelayer(s.ctx, s.eRep, s.initialPath))
	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.stargaze, s.slothchain))

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

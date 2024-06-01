package interslothtest

import (
	"testing"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"
)

type SmokeTestSuite struct {
	E2ETestSuite
}

func TestSmokeTestSuite(t *testing.T) {
	testifysuite.Run(t, new(SmokeTestSuite))
}

func (s *SmokeTestSuite) TestChainStarts() {
	s.NotNil(s.ic)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain))
}

func (s *SmokeTestSuite) TestIBCTokenTransfers() {
	slothUser, err := s.slothchain.BuildWallet(s.ctx, "slothUser", "")
	s.NoError(err)
	stargazeUser := interchaintest.GetAndFundTestUsers(s.T(), s.ctx, s.T().Name(), math.NewInt(10_000_000_000), s.stargaze)[0]

	s.NoError(s.r.StartRelayer(s.ctx, s.eRep, s.initialPath))

	s.T().Cleanup(
		func() {
			err := s.r.StopRelayer(s.ctx, s.eRep)
			if err != nil {
				s.T().Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze))

	channel, err := ibc.GetTransferChannel(s.ctx, s.r, s.eRep, s.stargaze.Config().ChainID, s.slothchain.Config().ChainID)
	s.NoError(err)

	sgUserAddr := stargazeUser.FormattedAddress()
	slothUserAddr := slothUser.FormattedAddress()

	sgBalanceBefore, err := s.stargaze.GetBalance(s.ctx, sgUserAddr, s.stargaze.Config().Denom)
	s.NoError(err)

	var transferAmount = math.NewInt(1_000)
	transfer := ibc.WalletAmount{
		Address: slothUserAddr,
		Denom:   s.stargaze.Config().Denom,
		Amount:  transferAmount,
	}
	_, err = s.stargaze.SendIBCTransfer(s.ctx, channel.ChannelID, sgUserAddr, transfer, ibc.TransferOptions{})
	s.NoError(err)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze))
	srcTokenDenom := transfertypes.GetPrefixedDenom("transfer", channel.Counterparty.ChannelID, s.stargaze.Config().Denom)
	srcIBCDenom := transfertypes.ParseDenomTrace(srcTokenDenom).IBCDenom()

	sgBalanceAfter, err := s.stargaze.GetBalance(s.ctx, sgUserAddr, s.stargaze.Config().Denom)
	s.NoError(err)
	s.Equal(sgBalanceBefore.Sub(transferAmount), sgBalanceAfter)

	slothBalanceAfter, err := s.slothchain.GetBalance(s.ctx, slothUserAddr, srcIBCDenom)
	s.NoError(err)
	s.Equal(transferAmount, slothBalanceAfter)

	// Transfer back
	transfer = ibc.WalletAmount{
		Address: sgUserAddr,
		Denom:   srcIBCDenom,
		Amount:  transferAmount,
	}
	_, err = s.slothchain.SendIBCTransfer(s.ctx, channel.Counterparty.ChannelID, slothUserAddr, transfer, ibc.TransferOptions{})
	s.NoError(err)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze))

	sgBalanceFinal, err := s.stargaze.GetBalance(s.ctx, sgUserAddr, s.stargaze.Config().Denom)
	s.NoError(err)
	s.Equal(sgBalanceBefore, sgBalanceFinal)

	slothBalanceFinal, err := s.slothchain.GetBalance(s.ctx, slothUserAddr, srcIBCDenom)
	s.NoError(err)
	s.Equal(math.NewInt(0), slothBalanceFinal)
}

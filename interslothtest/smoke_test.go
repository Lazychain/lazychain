package interslothtest

import (
	"encoding/hex"
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

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze, s.celestia))
}

func (s *SmokeTestSuite) TestIBCTokenTransfers() {
	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze, s.celestia))

	slothUser, err := s.slothchain.BuildWallet(s.ctx, "slothUser", "")
	s.NoError(err)
	stargazeUser := interchaintest.GetAndFundTestUsers(s.T(), s.ctx, s.T().Name(), math.NewInt(10_000_000_000), s.stargaze)[0]
	celestiaUser := interchaintest.GetAndFundTestUsers(s.T(), s.ctx, s.T().Name(), math.NewInt(10_000_000_000), s.celestia)[0]

	s.NoError(s.r.StartRelayer(s.ctx, s.eRep, s.sgSlothPath, s.celestiaSlothPath))

	s.T().Cleanup(
		func() {
			err := s.r.StopRelayer(s.ctx, s.eRep)
			if err != nil {
				s.T().Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze, s.celestia))

	sgToSlothChannel, err := ibc.GetTransferChannel(s.ctx, s.r, s.eRep, s.stargaze.Config().ChainID, s.slothchain.Config().ChainID)
	s.NoError(err)

	celestiaToSlothChannel, err := ibc.GetTransferChannel(s.ctx, s.r, s.eRep, s.celestia.Config().ChainID, s.slothchain.Config().ChainID)
	s.NoError(err)

	sgUserAddr := stargazeUser.FormattedAddress()
	celestiaUserAddr := celestiaUser.FormattedAddress()
	slothUserAddr := slothUser.FormattedAddress()

	sgBalanceBefore, err := s.stargaze.GetBalance(s.ctx, sgUserAddr, s.stargaze.Config().Denom)
	s.NoError(err)

	celestiaBalanceBefore, err := s.celestia.GetBalance(s.ctx, celestiaUserAddr, s.celestia.Config().Denom)
	s.NoError(err)

	var transferAmount = math.NewInt(1_000)
	sgTransfer := ibc.WalletAmount{
		Address: slothUserAddr,
		Denom:   s.stargaze.Config().Denom,
		Amount:  transferAmount,
	}
	_, err = s.stargaze.SendIBCTransfer(s.ctx, sgToSlothChannel.ChannelID, sgUserAddr, sgTransfer, ibc.TransferOptions{})
	s.NoError(err)

	celestiaTransfer := ibc.WalletAmount{
		Address: slothUserAddr,
		Denom:   s.celestia.Config().Denom,
		Amount:  transferAmount,
	}
	// Different versions makes the helper methods fail, so the celestia transfer is done more manually:
	txHash, err := s.celestia.GetNode().SendIBCTransfer(s.ctx, celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer, ibc.TransferOptions{})
	s.NoError(err)
	rpcNode, err := s.celestia.GetNode().CliContext().GetNode()
	s.NoError(err)
	hash, err := hex.DecodeString(txHash)
	s.NoError(err)
	resTx, err := rpcNode.Tx(s.ctx, hash, false)
	s.NoError(err)
	s.Equal(uint32(0), resTx.TxResult.Code)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze, s.celestia))

	starsSrcDenom := transfertypes.GetPrefixedDenom("transfer", sgToSlothChannel.Counterparty.ChannelID, s.stargaze.Config().Denom)
	starsSrcIBCDenom := transfertypes.ParseDenomTrace(starsSrcDenom).IBCDenom()
	tiaSrcDenom := transfertypes.GetPrefixedDenom("transfer", celestiaToSlothChannel.Counterparty.ChannelID, s.celestia.Config().Denom)
	tiaSrcIBCDenom := transfertypes.ParseDenomTrace(tiaSrcDenom).IBCDenom()

	sgBalanceAfter, err := s.stargaze.GetBalance(s.ctx, sgUserAddr, s.stargaze.Config().Denom)
	s.NoError(err)
	s.Equal(sgBalanceBefore.Sub(transferAmount), sgBalanceAfter)

	celestiaBalanceAfter, err := s.celestia.GetBalance(s.ctx, celestiaUserAddr, s.celestia.Config().Denom)
	s.NoError(err)
	s.Equal(celestiaBalanceBefore.Sub(transferAmount), celestiaBalanceAfter)

	slothStarsBalanceAfter, err := s.slothchain.GetBalance(s.ctx, slothUserAddr, starsSrcIBCDenom)
	s.NoError(err)
	s.Equal(transferAmount, slothStarsBalanceAfter)

	slothTiaBalanceAfter, err := s.slothchain.GetBalance(s.ctx, slothUserAddr, tiaSrcIBCDenom)
	s.NoError(err)
	s.Equal(transferAmount, slothTiaBalanceAfter)

	// Transfer back
	sgTransfer = ibc.WalletAmount{
		Address: sgUserAddr,
		Denom:   starsSrcIBCDenom,
		Amount:  transferAmount,
	}
	_, err = s.slothchain.SendIBCTransfer(s.ctx, sgToSlothChannel.Counterparty.ChannelID, slothUserAddr, sgTransfer, ibc.TransferOptions{})
	s.NoError(err)

	celestiaTransfer = ibc.WalletAmount{
		Address: celestiaUserAddr,
		Denom:   tiaSrcIBCDenom,
		Amount:  transferAmount,
	}
	_, err = s.slothchain.SendIBCTransfer(s.ctx, celestiaToSlothChannel.Counterparty.ChannelID, slothUserAddr, celestiaTransfer, ibc.TransferOptions{})
	s.NoError(err)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze))

	sgBalanceFinal, err := s.stargaze.GetBalance(s.ctx, sgUserAddr, s.stargaze.Config().Denom)
	s.NoError(err)
	s.Equal(sgBalanceBefore, sgBalanceFinal)

	celestiaBalanceFinal, err := s.celestia.GetBalance(s.ctx, celestiaUserAddr, s.celestia.Config().Denom)
	s.NoError(err)
	s.Equal(celestiaBalanceBefore, celestiaBalanceFinal)

	slothStarsBalanceFinal, err := s.slothchain.GetBalance(s.ctx, slothUserAddr, starsSrcIBCDenom)
	s.NoError(err)
	s.Equal(math.NewInt(0), slothStarsBalanceFinal)

	slothTiaBalanceFinal, err := s.slothchain.GetBalance(s.ctx, slothUserAddr, tiaSrcIBCDenom)
	s.NoError(err)
	s.Equal(math.NewInt(0), slothTiaBalanceFinal)
}

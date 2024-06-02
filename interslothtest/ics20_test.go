package interslothtest

import (
	"fmt"
	"go.uber.org/zap/zaptest"
	"testing"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	testifysuite "github.com/stretchr/testify/suite"
)

type ICS20TestSuite struct {
	E2ETestSuite
}

func TestICS20TestSuite(t *testing.T) {
	testifysuite.Run(t, new(ICS20TestSuite))
}

func (s *ICS20TestSuite) TestIBCTokenTransfers() {
	s.NotNil(s.ic)

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
	s.CelestiaIBCTransfer(celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer)

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

func (s *ICS20TestSuite) TestTIAGasToken() {
	celestiaUser := interchaintest.GetAndFundTestUsers(s.T(), s.ctx, s.T().Name(), math.NewInt(10_000_000_000), s.celestia)[0]
	slothUser, err := s.slothchain.BuildWallet(s.ctx, "slothUser", "")
	s.NoError(err)

	// Transfer TIA to relayer wallet + our user
	relayerWallet, found := s.r.GetWallet(s.slothchain.Config().ChainID)
	s.Require().True(found)

	s.NoError(s.r.StartRelayer(s.ctx, s.eRep, s.sgSlothPath, s.celestiaSlothPath))
	s.T().Cleanup(
		func() {
			err := s.r.StopRelayer(s.ctx, s.eRep)
			if err != nil {
				s.T().Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	celestiaToSlothChannel, err := ibc.GetTransferChannel(s.ctx, s.r, s.eRep, s.celestia.Config().ChainID, s.slothchain.Config().ChainID)
	s.NoError(err)

	var transferAmount = math.NewInt(1_000_000_00)
	celestiaTransfer := ibc.WalletAmount{
		Address: relayerWallet.FormattedAddress(),
		Denom:   s.celestia.Config().Denom,
		Amount:  transferAmount,
	}
	s.CelestiaIBCTransfer(celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer)

	celestiaTransfer = ibc.WalletAmount{
		Address: slothUser.FormattedAddress(),
		Denom:   s.celestia.Config().Denom,
		Amount:  transferAmount,
	}
	s.CelestiaIBCTransfer(celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer)

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze, s.celestia))

	// Change minimum gas price
	s.NoError(s.slothchain.StopAllNodes(s.ctx))
	for _, n := range s.slothchain.Nodes() {
		s.NoError(testutil.ModifyTomlConfigFile(
			s.ctx,
			zaptest.NewLogger(s.T()),
			n.DockerClient,
			s.T().Name(),
			n.VolumeName,
			"config/app.toml",
			testutil.Toml{
				"minimum-gas-prices": "0.025ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
			},
		))
	}
	s.NoError(s.slothchain.StartAllNodes(s.ctx))

	s.NoError(testutil.WaitForBlocks(s.ctx, 5, s.slothchain, s.stargaze, s.celestia))

	newWallet, err := s.slothchain.BuildWallet(s.ctx, "newWallet", "")
	s.NoError(err)
	_, err = s.slothchain.GetNode().ExecTx(s.ctx,
		slothUser.KeyName(), "bank", "send", slothUser.KeyName(), newWallet.FormattedAddress(),
		"42000000ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
		"--gas-prices", "0.025ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
	)
	s.NoError(err)

	slothUserBal, err := s.slothchain.GetBalance(s.ctx, slothUser.FormattedAddress(), "ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA")
	s.NoError(err)
	newWalletBal, err := s.slothchain.GetBalance(s.ctx, newWallet.FormattedAddress(), "ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA")
	s.NoError(err)

	fmt.Println(slothUserBal.String())

	s.Equal(math.NewInt(42000000), newWalletBal)
	// Because gas
	s.Less(slothUserBal.Int64(), int64(58_000_000))
	s.Greater(slothUserBal.Int64(), int64(57_000_000))
}

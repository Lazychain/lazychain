package ics20

import (
	"fmt"
	"github.com/Lazychain/lazychain/interchaintest/utils"
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
	utils.E2ETestSuite
}

func TestICS20TestSuite(t *testing.T) {
	testifysuite.Run(t, new(ICS20TestSuite))
}

func (s *ICS20TestSuite) TestIBCTokenTransfers() {
	s.NotNil(s.Interchain)

	slothUser, err := s.LazyChain.BuildWallet(s.Ctx, "slothUser", "")
	s.NoError(err)
	stargazeUser := interchaintest.GetAndFundTestUsers(s.T(), s.Ctx, s.T().Name(), math.NewInt(10_000_000_000), s.Stargaze)[0]
	celestiaUser := interchaintest.GetAndFundTestUsers(s.T(), s.Ctx, s.T().Name(), math.NewInt(10_000_000_000), s.Celestia)[0]

	s.NoError(s.Relayer.StartRelayer(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath, s.CelestiaSlothPath))
	s.T().Cleanup(
		func() {
			err := s.Relayer.StopRelayer(s.Ctx, s.RelayerExecRep)
			if err != nil {
				s.T().Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.LazyChain, s.Stargaze, s.Celestia))

	sgToSlothChannel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.RelayerExecRep, s.Stargaze.Config().ChainID, s.LazyChain.Config().ChainID)
	s.NoError(err)

	celestiaToSlothChannel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.RelayerExecRep, s.Celestia.Config().ChainID, s.LazyChain.Config().ChainID)
	s.NoError(err)

	sgUserAddr := stargazeUser.FormattedAddress()
	celestiaUserAddr := celestiaUser.FormattedAddress()
	slothUserAddr := slothUser.FormattedAddress()

	sgBalanceBefore, err := s.Stargaze.GetBalance(s.Ctx, sgUserAddr, s.Stargaze.Config().Denom)
	s.NoError(err)

	celestiaBalanceBefore, err := s.Celestia.GetBalance(s.Ctx, celestiaUserAddr, s.Celestia.Config().Denom)
	s.NoError(err)

	var transferAmount = math.NewInt(1_000)
	sgTransfer := ibc.WalletAmount{
		Address: slothUserAddr,
		Denom:   s.Stargaze.Config().Denom,
		Amount:  transferAmount,
	}
	_, err = s.Stargaze.SendIBCTransfer(s.Ctx, sgToSlothChannel.ChannelID, sgUserAddr, sgTransfer, ibc.TransferOptions{})
	s.NoError(err)

	celestiaTransfer := ibc.WalletAmount{
		Address: slothUserAddr,
		Denom:   s.Celestia.Config().Denom,
		Amount:  transferAmount,
	}
	s.CelestiaIBCTransfer(celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer)

	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.LazyChain, s.Stargaze, s.Celestia))

	starsSrcDenom := transfertypes.GetPrefixedDenom("transfer", sgToSlothChannel.Counterparty.ChannelID, s.Stargaze.Config().Denom)
	starsSrcIBCDenom := transfertypes.ParseDenomTrace(starsSrcDenom).IBCDenom()
	tiaSrcDenom := transfertypes.GetPrefixedDenom("transfer", celestiaToSlothChannel.Counterparty.ChannelID, s.Celestia.Config().Denom)
	tiaSrcIBCDenom := transfertypes.ParseDenomTrace(tiaSrcDenom).IBCDenom()

	sgBalanceAfter, err := s.Stargaze.GetBalance(s.Ctx, sgUserAddr, s.Stargaze.Config().Denom)
	s.NoError(err)
	s.Equal(sgBalanceBefore.Sub(transferAmount), sgBalanceAfter)

	celestiaBalanceAfter, err := s.Celestia.GetBalance(s.Ctx, celestiaUserAddr, s.Celestia.Config().Denom)
	s.NoError(err)
	s.Equal(celestiaBalanceBefore.Sub(transferAmount), celestiaBalanceAfter)

	slothStarsBalanceAfter, err := s.LazyChain.GetBalance(s.Ctx, slothUserAddr, starsSrcIBCDenom)
	s.NoError(err)
	s.Equal(transferAmount, slothStarsBalanceAfter)

	slothTiaBalanceAfter, err := s.LazyChain.GetBalance(s.Ctx, slothUserAddr, tiaSrcIBCDenom)
	s.NoError(err)
	s.Equal(transferAmount, slothTiaBalanceAfter)

	// Transfer back
	sgTransfer = ibc.WalletAmount{
		Address: sgUserAddr,
		Denom:   starsSrcIBCDenom,
		Amount:  transferAmount,
	}
	_, err = s.LazyChain.SendIBCTransfer(s.Ctx, sgToSlothChannel.Counterparty.ChannelID, slothUserAddr, sgTransfer, ibc.TransferOptions{})
	s.NoError(err)

	celestiaTransfer = ibc.WalletAmount{
		Address: celestiaUserAddr,
		Denom:   tiaSrcIBCDenom,
		Amount:  transferAmount,
	}
	_, err = s.LazyChain.SendIBCTransfer(s.Ctx, celestiaToSlothChannel.Counterparty.ChannelID, slothUserAddr, celestiaTransfer, ibc.TransferOptions{})
	s.NoError(err)

	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.LazyChain, s.Stargaze))

	sgBalanceFinal, err := s.Stargaze.GetBalance(s.Ctx, sgUserAddr, s.Stargaze.Config().Denom)
	s.NoError(err)
	s.Equal(sgBalanceBefore, sgBalanceFinal)

	celestiaBalanceFinal, err := s.Celestia.GetBalance(s.Ctx, celestiaUserAddr, s.Celestia.Config().Denom)
	s.NoError(err)
	s.Equal(celestiaBalanceBefore, celestiaBalanceFinal)

	slothStarsBalanceFinal, err := s.LazyChain.GetBalance(s.Ctx, slothUserAddr, starsSrcIBCDenom)
	s.NoError(err)
	s.Equal(math.NewInt(0), slothStarsBalanceFinal)

	slothTiaBalanceFinal, err := s.LazyChain.GetBalance(s.Ctx, slothUserAddr, tiaSrcIBCDenom)
	s.NoError(err)
	s.Equal(math.NewInt(0), slothTiaBalanceFinal)
}

func (s *ICS20TestSuite) TestTIAGasToken() {
	s.NotNil(s.Interchain)

	celestiaUser := interchaintest.GetAndFundTestUsers(s.T(), s.Ctx, s.T().Name(), math.NewInt(10_000_000_000), s.Celestia)[0]
	slothUser, err := s.LazyChain.BuildWallet(s.Ctx, "slothUser", "")
	s.NoError(err)

	// Transfer TIA to relayer wallet + our user
	relayerWallet, found := s.Relayer.GetWallet(s.LazyChain.Config().ChainID)
	s.Require().True(found)

	s.NoError(s.Relayer.StartRelayer(s.Ctx, s.RelayerExecRep, s.StargazeSlothPath, s.CelestiaSlothPath))
	s.T().Cleanup(
		func() {
			err := s.Relayer.StopRelayer(s.Ctx, s.RelayerExecRep)
			if err != nil {
				s.T().Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	celestiaToSlothChannel, err := ibc.GetTransferChannel(s.Ctx, s.Relayer, s.RelayerExecRep, s.Celestia.Config().ChainID, s.LazyChain.Config().ChainID)
	s.NoError(err)

	var transferAmount = math.NewInt(1_000_000_00)
	celestiaTransfer := ibc.WalletAmount{
		Address: relayerWallet.FormattedAddress(),
		Denom:   s.Celestia.Config().Denom,
		Amount:  transferAmount,
	}
	s.CelestiaIBCTransfer(celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer)

	celestiaTransfer = ibc.WalletAmount{
		Address: slothUser.FormattedAddress(),
		Denom:   s.Celestia.Config().Denom,
		Amount:  transferAmount,
	}
	s.CelestiaIBCTransfer(celestiaToSlothChannel.ChannelID, celestiaUser.KeyName(), celestiaTransfer)

	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.LazyChain, s.Stargaze, s.Celestia))

	// Change minimum gas price
	s.NoError(s.LazyChain.StopAllNodes(s.Ctx))
	for _, n := range s.LazyChain.Nodes() {
		s.NoError(testutil.ModifyTomlConfigFile(
			s.Ctx,
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
	s.NoError(s.LazyChain.StartAllNodes(s.Ctx))

	s.NoError(testutil.WaitForBlocks(s.Ctx, 5, s.LazyChain, s.Stargaze, s.Celestia))

	newWallet, err := s.LazyChain.BuildWallet(s.Ctx, "newWallet", "")
	s.NoError(err)
	_, err = s.LazyChain.GetNode().ExecTx(s.Ctx,
		slothUser.KeyName(), "bank", "send", slothUser.KeyName(), newWallet.FormattedAddress(),
		"42000000ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
		"--gas-prices", "0.025ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA",
	)
	s.NoError(err)

	slothUserBal, err := s.LazyChain.GetBalance(s.Ctx, slothUser.FormattedAddress(), "ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA")
	s.NoError(err)
	newWalletBal, err := s.LazyChain.GetBalance(s.Ctx, newWallet.FormattedAddress(), "ibc/C3E53D20BC7A4CC993B17C7971F8ECD06A433C10B6A96F4C4C3714F0624C56DA")
	s.NoError(err)

	fmt.Println(slothUserBal.String())

	s.Equal(math.NewInt(42000000), newWalletBal)
	// Because gas
	s.Less(slothUserBal.Int64(), int64(58_000_000))
	s.Greater(slothUserBal.Int64(), int64(57_000_000))
}

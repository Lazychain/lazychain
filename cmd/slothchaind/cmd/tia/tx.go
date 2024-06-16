package tia

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/gjermundgaraba/slothchain/cmd/slothchaind/cmd/lazycommandutils"
	"github.com/spf13/cobra"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "tia",
		Short:              "Transaction commands for TIA",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	cmd.AddCommand(TransferCmd())

	return cmd
}

func TransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [from_address] [to_address] [amount]",
		Short: "Transfer TIA tokens between celestia and slothchain (both ways)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			from := args[0]
			to := args[1]
			amountStr := args[2]

			node, _ := cmd.Flags().GetString(sdkflags.FlagNode)
			chainID, _ := cmd.Flags().GetString(sdkflags.FlagChainID)
			waitForTx, _ := cmd.Flags().GetBool(lazycommandutils.FlagWaitForTx)
			denom, _ := cmd.Flags().GetString(lazycommandutils.FlagICS20Denom)
			ics20Channel, _ := cmd.Flags().GetString(lazycommandutils.FlagICS20Channel)

			mainnet, _ := cmd.Flags().GetBool(lazycommandutils.FlagMainnet)
			testnet, _ := cmd.Flags().GetBool(lazycommandutils.FlagTestnet)

			isCelestia := strings.HasPrefix(from, "celestia")
			if isCelestia && !strings.HasPrefix(to, "lazy") {
				return fmt.Errorf("invalid addresses. Must transfer between celestia and slothchain")
			}
			if !isCelestia && (!strings.HasPrefix(to, "celestia") || !strings.HasPrefix(from, "lazy")) {
				return fmt.Errorf("invalid addresses. Must transfer between celestia and slothchain")
			}

			if !mainnet && !testnet &&
				(node == "" || chainID == "" || denom == "") {
				return fmt.Errorf("missing required flags. Either set --mainnet or --testnet or provide the manual flags (--%s --%s --%s --%s)",
					sdkflags.FlagNode, sdkflags.FlagChainID, lazycommandutils.FlagICS20Denom, lazycommandutils.FlagICS20Channel)
			}

			if mainnet {
				return fmt.Errorf("mainnet not supported yet")
			} else if testnet {
				var networkInfo lazycommandutils.StaticICS20NetworkInfo
				if isCelestia {
					networkInfo = lazycommandutils.ICS20Testnets.Celestia
				} else {
					networkInfo = lazycommandutils.ICS20Testnets.Slothchain
				}

				denom = networkInfo.ICS20Denom
				ics20Channel = networkInfo.ICS20Channel
				node = networkInfo.Node
				// Needed because this flag is picked up later by the clientCtx
				if err := cmd.Flags().Set(sdkflags.FlagNode, node); err != nil {
					return err
				}

				chainID = networkInfo.ChainID
				if err := cmd.Flags().Set(sdkflags.FlagChainID, chainID); err != nil {
					return err
				}
			}

			amount, err := strconv.ParseInt(amountStr, 10, 64)
			if err != nil {
				// Try to parse as coin
				coin, err := sdk.ParseCoinNormalized(amountStr)
				if err != nil {
					return err
				}

				amount = coin.Amount.Int64()
				denom = coin.Denom
			}

			now := time.Now()
			fiveMinutesLater := now.Add(5 * time.Minute) // TODO: Maybe more...
			msg := transfertypes.NewMsgTransfer("transfer", ics20Channel, sdk.NewCoin(denom, sdkmath.NewInt(amount)), from, to, clienttypes.Height{}, uint64(fiveMinutesLater.UnixNano()), "")

			if err := cmd.Flags().Set(sdkflags.FlagFrom, from); err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if waitForTx {
				if err := lazycommandutils.SendAndWaitForTx(clientCtx, cmd.Flags(), msg); err != nil {
					return err
				}

				fmt.Printf("🦥 lazy... transfer... of... %d%s... to... %s... done...\n", amount, denom, to)
				fmt.Printf("🦥 tx... finally... done... time... too... 💤!\n")

				return nil
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(lazycommandutils.FlagWaitForTx, true, "Wait for transaction to be included in a block")
	cmd.Flags().Bool(lazycommandutils.FlagMainnet, false, "Use mainnet values")
	cmd.Flags().Bool(lazycommandutils.FlagTestnet, false, "Use testnet values")
	cmd.Flags().String(lazycommandutils.FlagICS20Denom, "", "Denom of ICS20 token on sender chain (optional if using --testnet or --mainnet)")

	sdkflags.AddTxFlagsToCmd(cmd)
	nodeFlag := cmd.Flags().Lookup(sdkflags.FlagNode)
	nodeFlag.Usage = "RPC endpoint of sending chain (celestia or slothchain)"
	nodeFlag.DefValue = ""

	cmd.Flags().Lookup(sdkflags.FlagChainID).Usage = "Chain ID of sending chain (stargaze or slothchain)"

	return cmd
}

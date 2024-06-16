package tia

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/gjermundgaraba/slothchain/cmd/slothchaind/cmd/lazycommandutils"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "tia",
		Short:              "Query commands for TIA",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	cmd.AddCommand(QueryTIABalanceCmd())

	return cmd
}

func QueryTIABalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance [address]",
		Short: "Query tia balance of an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			address := args[0]

			node, _ := cmd.Flags().GetString(sdkflags.FlagNode)
			denom, _ := cmd.Flags().GetString(lazycommandutils.FlagICS20Denom)
			mainnet, _ := cmd.Flags().GetBool(lazycommandutils.FlagMainnet)
			testnet, _ := cmd.Flags().GetBool(lazycommandutils.FlagTestnet)

			if !mainnet && !testnet && node == "" {
				return fmt.Errorf("missing required flags. Either set --mainnet or --testnet or provide the manual flags (--%s --%s)",
					sdkflags.FlagNode, lazycommandutils.FlagNFTContract)
			}

			isCelestia := strings.HasPrefix(address, "celestia")
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
				node = networkInfo.Node
				// Needed because this flag is picked up later by the clientCtx
				if err := cmd.Flags().Set(sdkflags.FlagNode, node); err != nil {
					return err
				}
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := banktypes.NewQueryClient(clientCtx)
			accAddr, err := sdk.AccAddressFromBech32(address)
			if err != nil {
				return err
			}
			params := banktypes.NewQueryBalanceRequest(accAddr, denom)
			res, err := queryClient.Balance(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Balance)
		},
	}

	cmd.Flags().Bool(lazycommandutils.FlagMainnet, false, "Use mainnet values")
	cmd.Flags().Bool(lazycommandutils.FlagTestnet, false, "Use testnet values")
	cmd.Flags().String(lazycommandutils.FlagICS20Denom, "", "Denom of ICS20 token (optional if using --testnet or --mainnet)")

	sdkflags.AddQueryFlagsToCmd(cmd)

	nodeFlag := cmd.Flags().Lookup(sdkflags.FlagNode)
	nodeFlag.DefValue = ""
	nodeFlag.Usage = "RPC endpoint of chain to query (celestia or slothchain)"

	return cmd
}

package tia

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "tia",
		Short:              "Transaction commands for tia",
		DisableFlagParsing: true,
		RunE:               client.ValidateCmd,
	}

	cmd.AddCommand(DepositCmd())

	return cmd
}

func DepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [from_key_or_address] [to_address]",
		Short: "Deposit tia tokens to celstia to slothchain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}

	return cmd
}

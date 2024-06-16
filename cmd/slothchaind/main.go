package main

import (
	"fmt"
	"os"
	"strings"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/gjermundgaraba/slothchain/app"
	"github.com/gjermundgaraba/slothchain/cmd/slothchaind/cmd"
)

func main() {
	// Hack to be able to use the same binary for custom commands that send txs to stargaze
	if prefixOverride := getAddressPrefixOverride(); prefixOverride != "" {
		app.AccountAddressPrefix = prefixOverride
	}

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		_, _ = fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func getAddressPrefixOverride() string {
	args := os.Args[1:]
	if len(args) > 3 &&
		args[1] == "sloths" &&
		(args[2] == "transfer" || args[2] == "owned-by") {
		if strings.HasPrefix(args[3], "stars") {
			return "stars"
		}
	}

	if len(args) > 3 &&
		args[1] == "tia" &&
		(args[2] == "transfer" || args[2] == "balance") {
		if strings.HasPrefix(args[3], "celestia") {
			return "celestia"
		}
	}

	return ""
}

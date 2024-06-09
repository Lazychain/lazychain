package main

import (
	"fmt"
	"os"
	"strings"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"slothchain/app"
	"slothchain/cmd/slothchaind/cmd"
)

func main() {
	// Hack to be able to use the same binary for custom commands that send txs to stargaze
	if isStargaze() {
		app.AccountAddressPrefix = "stars"
	}

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func isStargaze() bool {
	args := os.Args[1:]
	if len(args) > 3 && args[1] == "sloths" && (args[2] == "transfer" || args[2] == "owned-by") {
		if strings.HasPrefix(args[3], "stars") {
			return true
		}
	}

	return false
}

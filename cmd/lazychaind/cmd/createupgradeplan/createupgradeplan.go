package createupgradeplan

import (
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"encoding/json"
	"fmt"
	sdkflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func CreateUpgradePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-upgrade-plan [name] [height]",
		Short: "Create an upgrade plan file",
		Long:  `Create an upgrade plan file that will be picked up by app.UpgradeKeeper.ReadUpgradeInfoFromDisk() on startup`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			height, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			plan := upgradetypes.Plan{
				Name:   name,
				Height: height,
				Info:   "",
			}

			homeDir, _ := cmd.Flags().GetString(sdkflags.FlagHome)
			if homeDir == "" {
				return fmt.Errorf("home directory must be specified")
			}

			upgradeInfoFileDir := path.Join(homeDir, "data")
			if err := os.MkdirAll(upgradeInfoFileDir, os.ModePerm); err != nil {
				return fmt.Errorf("could not create directory %q: %w", upgradeInfoFileDir, err)
			}
			fullUpgradeInfoPath := filepath.Join(upgradeInfoFileDir, upgradetypes.UpgradeInfoFilename)

			jsonBz, err := json.Marshal(plan)
			if err != nil {
				return err
			}
			if err := os.WriteFile(fullUpgradeInfoPath, jsonBz, 0644); err != nil {
				return err
			}

			fmt.Printf("🦥 Created upgrade plan file %q\n", fullUpgradeInfoPath)

			return nil
		},
	}

	return cmd
}

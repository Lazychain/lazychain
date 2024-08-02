package v1_1

import (
	"context"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/Lazychain/lazychain/app/upgrades"
	tokenfactorytypes "github.com/Stride-Labs/tokenfactory/tokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const upgradeName = "v1.1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          upgradeName,
	CreateUpgradeHandler: createUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{tokenfactorytypes.StoreKey},
	},
}

// CreateUpgradeHandler creates an SDK upgrade handler for v1.1
func createUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		sdkCtx.Logger().Info(`    _    _ _____   _____ _____            _____  ______      `)
		sdkCtx.Logger().Info(`   | |  | |  __ \ / ____|  __ \     /\   |  __ \|  ____|     `)
		sdkCtx.Logger().Info(`   | |  | | |__) | |  __| |__) |   /  \  | |  | | |__        `)
		sdkCtx.Logger().Info(`   | |  | |  ___/| | |_ |  _  /   / /\ \ | |  | |  __|       `)
		sdkCtx.Logger().Info(`   | |__| | |    | |__| | | \ \  / ____ \| |__| | |____      `)
		sdkCtx.Logger().Info(`    \____/|_|     \_____|_|  \_\/_/    \_\_____/|______|     `)
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣶⠖⢶⣦⣄⡀⠀⢀⣴⣶⠟⠓⣶⣦⣄⡀⠀⠀⠀⠀⠀⣀⣤⣤⣀⡀⠀⠀⢠⣤⣤⣄⡀⠀⠀⠀⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡿⣿⡄⠀⣿⠈⢻⣤⣾⠏⠀⠀⠀⠈⢷⡈⠻⣦⡀⠀⣠⣾⠟⠋⠀⠙⣿⣶⣴⠏⢠⣿⠋⠉⣷⡄⠀⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⣿⠁⠈⣿⠀⣿⠃⢸⣿⡏⠀⠀⠀⠀⠀⢸⡻⣦⣹⣇⢠⣿⠃⠀⠀⠀⠀⠘⣇⠙⣧⡿⢁⣴⠞⠛⢿⡆⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⢰⣿⠃⠀⠀⢹⣷⣿⣰⡿⢿⡇⠀⠀⠀⠀⠀⢸⢻⣜⣷⣿⣿⡇⠀⠀⠀⠀⠀⠀⣟⣷⣹⣁⡞⠁⠀⠀⠘⣿⡄⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⣿⡟⠀⠀⠀⠈⠉⢹⡟⠁⢸⡇⠀⠀⠀⠀⠀⢸⡆⠙⠃⢀⣿⠀⠀⠀⠀⠀⠀⠀⣿⠛⣿⠛⠀⠀⠀⠀⠀⢹⡇⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⠀⠈⣇⠀⢸⡇⠀⠀⠀⠀⠀⠀⣇⠀⠀⠘⣿⡄⠀⠀⠀⠀⠀⠀⢹⡀⢸⡇⠀⠀⠀⠀⠀⠘⣿⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⣿⡇⠀⠀⠀⠀⠀⠀⣿⡀⢸⣷⠀⠀⠀⠀⠀⠀⢻⡆⠀⠀⣿⡇⠀⠀⠀⠀⠀⠀⠸⣧⢸⣿⠀⠀⠀⠀⠀⠀⢿⡆")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⣠⣤⣶⣶⣿⠿⢶⣶⣤⣄⠀⠀⠘⡇⠀⢻⡄⠀⠀⠀⠀⠀⠀⢳⡀⠀⢻⣧⠀⠀⠀⠀⠀⠀⠀⢿⣾⣿⠀⠀⠀⠀⠀⠀⢸⡇")
		sdkCtx.Logger().Info("⠀⠀⣴⣿⢟⣯⣭⡎⠀⠀⢀⣤⡟⠻⣷⣄⠀⢹⡄⢸⣇⠀⠀⠀⠀⠀⠀⠈⣷⡀⠘⣿⡄⠀⠀⠀⠀⠀⠀⠀⢿⡏⠀⠀⠀⠀⠀⠀⢸⣧")
		sdkCtx.Logger().Info("⢀⣾⢏⠞⠉⠀⠀⠀⠀⠀⠻⢿⣿⣀⠼⢿⣶⣶⣷⡀⣿⡀⠀⠀⠀⠀⠀⠀⢸⣇⠀⢻⣷⠀⠀⠀⠀⠀⠀⠀⠈⢿⣆⠀⠀⠀⠀⠀⢸⣿")
		sdkCtx.Logger().Info("⣸⡷⠋⠀⠀⠀⠀⣠⣶⣿⠦⣤⠄⠀⠀⠀⣿⣿⣦⡀⣿⠇⠀⠀⠀⠀⠀⠀⠀⢻⡀⠈⣿⡆⠀⠀⠀⠀⠀⠀⠀⠈⢻⣆⠀⠀⠀⠀⢸⣿")
		sdkCtx.Logger().Info("⣿⢁⡴⣾⣿⡆⠀⠙⢛⣡⡾⠋⠀⠀⠀⢠⠇⠈⠛⣿⠏⠀⠀⠀⠀⠀⠀⠰⣄⠸⡗⠛⢻⣿⠀⠀⠀⠀⠀⠀⠀⠀⠈⢻⣆⠀⠀⠀⣼⣿")
		sdkCtx.Logger().Info("⢿⡞⠀⠈⢉⡇⠉⠉⠉⠉⠀⠀⠀⠀⣠⠊⠀⢀⡾⠋⠀⠀⠀⠀⠀⢀⡀⠀⣿⠳⠇⠀⢸⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢹⡄⠀⠀⣿⡇")
		sdkCtx.Logger().Info("⠸⣷⠀⢠⠎⠀⠀⠀⠀⠀⠀⣀⠴⠋⠀⠖⠚⠋⠀⠀⠀⠀⠀⢀⡄⢸⣷⡀⣿⠀⠀⠀⣸⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡇⠀⢰⣿⠇")
		sdkCtx.Logger().Info("⠀⢻⣷⣯⣀⣀⣀⣀⣠⠤⠚⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠛⠛⠿⣷⠇⠀⠀⢠⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⠀⣾⡟⠀")
		sdkCtx.Logger().Info("⠀⠈⢻⣷⡉⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⡿⠃⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⢻⣿⣆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⣿⠃⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠙⣿⣷⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣿⠇⠀⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠈⠻⣿⣷⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡿⠃⠀⠀⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠈⠛⢿⣿⣶⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣾⠟⠁⠀⠀⠀⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠻⢿⣿⣶⣦⣄⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣠⣴⣶⣿⠿⠋⠁⠀⠀⠀⠀⠀⠀⠀")
		sdkCtx.Logger().Info("⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠛⡻⠿⠿⣿⣿⣷⣶⣶⣶⣶⣶⣶⣶⣶⣶⣿⣿⠿⠿⢛⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀")

		sdkCtx.Logger().Info("Running module migrations for %s...", upgradeName)
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

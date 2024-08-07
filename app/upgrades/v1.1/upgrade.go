package v1_1

import (
	"context"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/Lazychain/lazychain/app/upgrades"
	tokenfactorytypes "github.com/Stride-Labs/tokenfactory/tokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	solomachine "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	sequencertypes "github.com/decentrio/rollkit-sdk/x/sequencer/types"
)

const upgradeName = "v1.1"

var added = []string{
	tokenfactorytypes.StoreKey,
	govtypes.StoreKey,
	sequencertypes.StoreKey,
}

var Upgrade = upgrades.Upgrade{
	UpgradeName:          upgradeName,
	CreateUpgradeHandler: createUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: added,
		Deleted: []string{
			"slashing",
			"evidence",
			"mint",
		},
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

		fromVM[wasmtypes.ModuleName] = 4
		fromVM[exported.ModuleName] = 6
		fromVM[ibctransfertypes.ModuleName] = 5
		fromVM[ibcfeetypes.ModuleName] = 2
		fromVM[capabilitytypes.ModuleName] = 1
		fromVM[ibctm.ModuleName] = 0
		fromVM[solomachine.ModuleName] = 0

		for _, toBeAdded := range added {
			if _, ok := fromVM[toBeAdded]; ok {
				panic("module already exists")
			}
		}

		for moduleName, ver := range fromVM {
			sdkCtx.Logger().Info("fromVM", "module", moduleName, "version", ver)
		}

		sdkCtx.Logger().Info("Running module migrations for %s...", upgradeName)
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

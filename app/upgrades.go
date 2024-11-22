package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/Lazychain/lazychain/app/upgrades"
)

var Upgrades = []upgrades.Upgrade{}

func (app *LazyApp) setupUpgradeHandlers() {
	// register upgrade handlers
	for _, upgradeDetails := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgradeDetails.UpgradeName,
			upgradeDetails.CreateUpgradeHandler(
				app.ModuleManager,
				app.Configurator(),
			),
		)
	}

	// register store loaders
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("Failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for i := range Upgrades {
		if upgradeInfo.Name == Upgrades[i].UpgradeName {
			app.BaseApp.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &Upgrades[i].StoreUpgrades))
		}
	}
}

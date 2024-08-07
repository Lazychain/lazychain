package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/Lazychain/lazychain/app/upgrades"
	v1_1 "github.com/Lazychain/lazychain/app/upgrades/v1.1"
)

var Upgrades = []upgrades.Upgrade{
	v1_1.Upgrade,
}

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

	upgradePlanPath, err := app.UpgradeKeeper.GetUpgradeInfoPath()
	if err != nil {
		panic(fmt.Sprintf("Failed to get upgrade info path %s", err))
	}
	app.Logger().Info("🦥 Looking for upgrade plan", "upgradePlanPath", upgradePlanPath)

	// register store loaders
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("Failed to read upgrade info from disk %s", err))
	}

	if upgradeInfo.Name != "" {
		app.Logger().Info("🦥 Upgrade plan found", "name", upgradeInfo.Name, "height", upgradeInfo.Height)
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

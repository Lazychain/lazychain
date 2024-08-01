package upgrades

import (
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// UpgradeName defines the on-chain upgrade name for the upgrade, e.g. "v1.8", "v1.9", etc.
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	// mm *module.Manager, computeModule *computetypes.AppModule, configurator module.Configurator
	CreateUpgradeHandler func(
		mm *module.Manager,
		configurator module.Configurator,
	) upgradetypes.UpgradeHandler

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades store.StoreUpgrades
}

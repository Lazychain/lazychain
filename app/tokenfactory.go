package app

import (
	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"

	"github.com/Stride-Labs/tokenfactory/tokenfactory"
	tokenfactorykeeper "github.com/Stride-Labs/tokenfactory/tokenfactory/keeper"
	tokenfactorytypes "github.com/Stride-Labs/tokenfactory/tokenfactory/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// registerTokenFactoryModule register the tokenfactory keeper and non dependency inject modules
func (app *LazyApp) registerTokenFactoryModule() error {
	// set up non depinject support modules store keys
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(tokenfactorytypes.StoreKey),
	); err != nil {
		panic(err)
	}

	// register the key tables for legacy param subspaces
	app.ParamsKeeper.Subspace(tokenfactorytypes.ModuleName)

	// Cast module account permissions
	maccPerms := make(map[string][]string)
	for _, modulePerms := range moduleAccPerms {
		maccPerms[modulePerms.Account] = modulePerms.Permissions
	}

	// Token factory keeper
	app.TokenFactoryKeeper = tokenfactorykeeper.NewKeeper(
		app.GetKey(tokenfactorytypes.StoreKey),
		app.GetSubspace(tokenfactorytypes.ModuleName),
		maccPerms,
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
	)

	// register tokenfactory module
	if err := app.RegisterModules(
		tokenfactory.NewAppModule(app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper)); err != nil {
		return err
	}

	return nil
}

// RegisterTokenFactory Since the tokenfactory doesn't support dependency injection,
// we need to manually register the modules on the client side.
func RegisterTokenFactory(registry cdctypes.InterfaceRegistry) (string, appmodule.AppModule) {
	name := tokenfactorytypes.ModuleName
	mod := tokenfactory.AppModule{}
	module.CoreAppModuleBasicAdaptor(name, mod).RegisterInterfaces(registry)
	return name, mod
}

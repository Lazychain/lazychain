package app

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/Stride-Labs/tokenfactory/tokenfactory"
	tokenfactorykeeper "github.com/Stride-Labs/tokenfactory/tokenfactory/keeper"
	tokenfactorytypes "github.com/Stride-Labs/tokenfactory/tokenfactory/types"
)

func (app *LazyApp) registerTokenFactoryModule() error {
	// set up non depinject support modules store keys
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(tokenfactorytypes.StoreKey),
	); err != nil {
		panic(err)
	}

	// Token factory keeper
	app.TokenFactoryKeeper = tokenfactorykeeper.NewKeeper(
		app.GetKey(tokenfactorytypes.StoreKey),
		app.GetSubspace(tokenfactorytypes.ModuleName),
		GetMaccPerms(),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
	)

	// register IBC modules
	if err := app.RegisterModules(
		tokenfactory.NewAppModule(app.TokenFactoryKeeper, app.AccountKeeper, app.BankKeeper)); err != nil {
		return err
	}

	return nil
}

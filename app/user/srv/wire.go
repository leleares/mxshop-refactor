//go:build wireinject
// +build wireinject

package srv

import (
	"github.com/google/wire"
	"mxshop/app/pkg/options"
	"mxshop/app/user/srv/controller"
	"mxshop/app/user/srv/data"
	db "mxshop/app/user/srv/data/db"
	"mxshop/app/user/srv/service"
	gapp "mxshop/gmicro/app"
	"mxshop/pkg/log"
)

func initApp(*log.Options, *options.ServerOptions, *options.RegistryOptions, *options.TelemetryOptions, *options.MySQLOptions) (*gapp.App, error) {
	wire.Build(ProviderSet, controller.ProviderSet, service.ProviderSet, data.ProviderSet, db.ProviderSet)
	return &gapp.App{}, nil
}

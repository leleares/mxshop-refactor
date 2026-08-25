package srv

import (
	"mxshop/app/goods/srv/config"
	"mxshop/app/pkg/options"
	gapp "mxshop/gmicro/app"
	"mxshop/gmicro/registry"
	"mxshop/gmicro/registry/consul"
	"mxshop/pkg/app"
	"mxshop/pkg/log"

	"github.com/hashicorp/consul/api"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("goods", basename, app.WithOptions(cfg), app.WithRunFunc(run(cfg)))
	return appl
}

func run(cfg *config.Config) app.RunFunc {
	return func(basename string) error {
		// goods grpc 服务
		goodsApp, err := NewGoodsApp(cfg)
		if err != nil {
			return err
		}
		// 启动 grpc 服务
		err = goodsApp.Run()
		if err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}

func NewGoodsApp(cfg *config.Config) (*gapp.App, error) {
	//初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	//服务注册
	register := NewRegistrar(cfg.Registry)

	//生成rpc服务
	rpcServer, err := NewGoodsRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	return gapp.New(
		gapp.WithName(cfg.Server.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func NewRegistrar(registry *options.RegistryOptions) registry.Registrar {
	c := api.DefaultConfig()
	c.Address = registry.Address
	c.Scheme = registry.Scheme
	cli, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

package srv

import (
	"context"

	"github.com/hashicorp/consul/api"
	"mxshop/app/inventory/srv/config"
	"mxshop/app/pkg/options"
	gapp "mxshop/gmicro/app"
	"mxshop/pkg/app"
	"mxshop/pkg/log"
	"mxshop/pkg/storage"

	"mxshop/gmicro/registry"
	"mxshop/gmicro/registry/consul"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("Inventory",
		"mxshop",
		app.WithOptions(cfg),
		app.WithRunFunc(run(cfg)),
		//app.WithNoConfig(), //设置不读取配置文件
	)
	return appl
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

func NewInventoryApp(cfg *config.Config) (*gapp.App, error) {
	//初始化log
	log.Init(cfg.Log)
	defer log.Flush()

	//服务注册
	register := NewRegistrar(cfg.Registry)

	//连接redis
	redisConfig := &storage.Config{
		Host:                  cfg.RedisOptions.Host,
		Port:                  cfg.RedisOptions.Port,
		Addrs:                 cfg.RedisOptions.Addrs,
		MasterName:            cfg.RedisOptions.MasterName,
		Username:              cfg.RedisOptions.Username,
		Password:              cfg.RedisOptions.Password,
		Database:              cfg.RedisOptions.Database,
		MaxIdle:               cfg.RedisOptions.MaxIdle,
		MaxActive:             cfg.RedisOptions.MaxActive,
		Timeout:               cfg.RedisOptions.Timeout,
		EnableCluster:         cfg.RedisOptions.EnableCluster,
		UseSSL:                cfg.RedisOptions.UseSSL,
		SSLInsecureSkipVerify: cfg.RedisOptions.SSLInsecureSkipVerify,
		EnableTracing:         cfg.RedisOptions.EnableTracing,
	}
	go storage.ConnectToRedis(context.Background(), redisConfig)

	//生成rpc服务
	rpcServer, err := NewInventoryRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	return gapp.New(
		gapp.WithName(cfg.Server.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		InventoryApp, err := NewInventoryApp(cfg)
		if err != nil {
			return err
		}

		//启动
		if err := InventoryApp.Run(); err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}

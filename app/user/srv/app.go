package srv

import (
	"mxshop/app/pkg/options"
	"mxshop/app/user/srv/config"
	gapp "mxshop/gmicro/app"
	"mxshop/gmicro/server/rpcserver"
	"mxshop/pkg/app"
	"mxshop/pkg/log"

	"github.com/hashicorp/consul/api"

	"mxshop/gmicro/registry"
	"mxshop/gmicro/registry/consul"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("user",
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

func NewUserApp(logOpts *log.Options, register registry.Registrar,
	serverOpts *options.ServerOptions, rpcServer *rpcserver.Server) (*gapp.App, error) {
	//初始化log
	log.Init(logOpts)
	defer log.Flush()

	return gapp.New(
		gapp.WithName(serverOpts.Name),
		gapp.WithRPCServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

// initApp 手动创建依赖注入（不使用 Wire）
func initApp(cfg *config.Config) (*gapp.App, error) {
	// 1. 创建 RPC 服务器（内部会创建 Data、Service、Controller 层）
	rpcServer, err := NewUserRPCServer(cfg)
	if err != nil {
		return nil, err
	}

	// 2. 创建服务注册器
	registrar := NewRegistrar(cfg.Registry)

	// 3. 创建应用
	userApp, err := NewUserApp(cfg.Log, registrar, cfg.Server, rpcServer)
	if err != nil {
		return nil, err
	}

	return userApp, nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		userApp, err := initApp(cfg)
		if err != nil {
			return err
		}

		//启动
		if err := userApp.Run(); err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}

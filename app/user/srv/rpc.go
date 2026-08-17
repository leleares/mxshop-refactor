package srv

import (
	"fmt"
	upb "mxshop/api/user/v1"
	"mxshop/app/pkg/options"
	"mxshop/gmicro/core/trace"
	"mxshop/gmicro/server/rpcserver"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"

	"mxshop/app/user/srv/config"
	"mxshop/app/user/srv/controller"
	"mxshop/app/user/srv/data"
	db "mxshop/app/user/srv/data/db"
	"mxshop/app/user/srv/service"

	"github.com/alibaba/sentinel-golang/pkg/datasource/nacos"
)

func NewUserRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})
	// 这里要进行 data、service、controller层的装配
	// 1. 初始化数据库连接
	gormDB, err := db.GetDBFactoryOr(cfg.MySQLOptions)
	if err != nil {
		return nil, err
	}
	userStore := data.NewUserStore(gormDB)
	// 2. 初始化service层
	userService := service.NewUserService(userStore)
	// 3. 初始化controller层
	userController := controller.NewUserController(userService)

	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	var opts []rpcserver.ServerOption
	opts = append(opts, rpcserver.WithAddress(rpcAddr))
	if cfg.Server.EnableLimit {
		// opts = append(opts, rpcserver.WithUnaryInterceptor(grpc.NewUnaryServerInterceptor()))
		// //我去初始化nacos
		// nds, err := NewNacosDataSource(cfg.NacosOptions)
		// if err != nil {
		// 	return nil, err
		// }
		// _ = nds
	}
	urpcServer := rpcserver.NewServer(opts...)

	upb.RegisterUserServer(urpcServer.Server, userController)

	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userController, r)
	//r.Run(":8075")
	return urpcServer, nil
}

func NewNacosDataSource(opts *options.NacosOptions) (*nacos.NacosDataSource, error) {
	//nacos server地址
	sc := []constant.ServerConfig{
		{
			ContextPath: "/nacos",
			Port:        opts.Port,
			IpAddr:      opts.Host,
		},
	}

	//nacos client 相关参数配置,具体配置可参考github.com/nacos-group/nacos-sdk-go
	cc := constant.ClientConfig{
		NamespaceId: opts.Namespace,
		TimeoutMs:   5000,
	}

	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": sc,
		"clientConfig":  cc,
	})
	if err != nil {
		return nil, err
	}

	//注册流控规则Handler
	h := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
	//创建NacosDataSource数据源
	nds, err := nacos.NewNacosDataSource(client, opts.Group, opts.DataId, h)
	if err != nil {
		return nil, err
	}
	return nds, nil
}

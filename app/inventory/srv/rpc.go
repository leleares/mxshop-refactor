package srv

import (
	"fmt"
	gpb "mxshop/api/inventory/v1"
	"mxshop/app/inventory/srv/config"
	v12 "mxshop/app/inventory/srv/internal/controller/v1"
	db2 "mxshop/app/inventory/srv/internal/data/v1/db"
	v13 "mxshop/app/inventory/srv/internal/service/v1"
	"mxshop/gmicro/core/trace"
	"mxshop/gmicro/server/rpcserver"

	"mxshop/pkg/log"
)

func NewInventoryRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	//有点繁琐，wire， ioc-golang
	dataFactory, err := db2.GetDBFactoryOr(cfg.MySQLOptions)
	if err != nil {
		log.Fatal(err.Error())
	}
	invService := v13.NewService(dataFactory, cfg.RedisOptions)
	invServer := v12.NewInventoryServer(invService)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))
	gpb.RegisterInventoryServer(grpcServer.Server, invServer)
	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userver, r)
	//r.Run(":8075")
	return grpcServer, nil
}

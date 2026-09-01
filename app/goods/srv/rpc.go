package srv

import (
	"fmt"
	gpb "mxshop/api/goods/v1"
	"mxshop/app/goods/srv/config"
	v12 "mxshop/app/goods/srv/controller"
	db "mxshop/app/goods/srv/data/db"
	es "mxshop/app/goods/srv/data_search/es"
	goodService "mxshop/app/goods/srv/service"
	"mxshop/gmicro/core/trace"
	"mxshop/gmicro/server/rpcserver"
)

func NewGoodsRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	// 初始化es工厂store
	esFactory, _ := es.GetSearchFactoryOr(cfg.EsOptions)
	// 初始化db工厂store
	dbFactory, _ := db.GetDBFactoryOr(cfg.MySQLOptions)
	goodSrv := goodService.NewService(dbFactory, esFactory)
	goodsServer := v12.NewGoodsServer(goodSrv)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))

	gpb.RegisterGoodsServer(grpcServer.Server, goodsServer)

	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userver, r)
	//r.Run(":8075")
	return grpcServer, nil
}

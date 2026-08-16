package srv

import (
	"fmt"
	gpb "mxshop/api/goods/v1"
	"mxshop/app/goods/srv/config"
	v12 "mxshop/app/goods/srv/internal/controller/v1"
	db2 "mxshop/app/goods/srv/internal/data/v1/db"
	"mxshop/app/goods/srv/internal/data_search/v1/es"
	v1 "mxshop/app/goods/srv/internal/service/v1"
	"mxshop/gmicro/core/trace"
	"mxshop/gmicro/server/rpcserver"

	"mxshop/pkg/log"
)

func NewGoodsRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
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

	//构建，繁琐 - 工厂模式
	searchFactory, err := es.GetSearchFactoryOr(cfg.EsOptions)
	if err != nil {
		log.Fatal(err.Error())
	}

	srvFactory := v1.NewService(dataFactory, searchFactory)
	goodsServer := v12.NewGoodsServer(srvFactory)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))

	gpb.RegisterGoodsServer(grpcServer.Server, goodsServer)

	//r := gin.Default()
	//upb.RegisterUserServerHTTPServer(userver, r)
	//r.Run(":8075")
	return grpcServer, nil
}

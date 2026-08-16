package srv

import (
	"fmt"
	gpb "mxshop/api/order/v1"
	"mxshop/app/order/srv/config"
	"mxshop/app/order/srv/internal/controller/order/v1"
	db2 "mxshop/app/order/srv/internal/data/v1/db"
	v13 "mxshop/app/order/srv/internal/service/v1"
	"mxshop/gmicro/core/trace"
	"mxshop/gmicro/server/rpcserver"

	"mxshop/pkg/log"
)

func NewOrderRPCServer(cfg *config.Config) (*rpcserver.Server, error) {
	//初始化open-telemetry的exporter
	trace.InitAgent(trace.Options{
		cfg.Telemetry.Name,
		cfg.Telemetry.Endpoint,
		cfg.Telemetry.Sampler,
		cfg.Telemetry.Batcher,
	})

	dataFactory, err := db2.GetDataFactoryOr(cfg.MySQLOptions, cfg.Registry)
	if err != nil {
		log.Fatal(err.Error())
	}

	orderSrvFactory := v13.NewService(dataFactory, cfg.Dtm)
	orderServer := order.NewOrderServer(orderSrvFactory)
	rpcAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	grpcServer := rpcserver.NewServer(rpcserver.WithAddress(rpcAddr))
	gpb.RegisterOrderServer(grpcServer.Server, orderServer)
	return grpcServer, nil
}

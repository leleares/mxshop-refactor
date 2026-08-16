package admin

import (
	"mxshop/app/user/srv/config"
	"mxshop/gmicro/server/restserver"
)

func NewUserHTTPServer(cfg *config.Config) (*restserver.Server, error) {
	urestServer := restserver.NewServer(restserver.WithPort(cfg.Server.HttpPort),
		restserver.WithMiddlewares(cfg.Server.Middlewares),
		restserver.WithMetrics(true),
	)

	//配置好路由
	initRouter(urestServer)

	return urestServer, nil
}

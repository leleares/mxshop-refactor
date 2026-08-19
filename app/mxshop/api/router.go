package admin

import (
	"mxshop/app/mxshop/api/config"
	userController "mxshop/app/mxshop/api/internal/controller/v1/user"
	userData "mxshop/app/mxshop/api/internal/data/rpc"
	userService "mxshop/app/mxshop/api/internal/service/v1/user"
	"mxshop/gmicro/server/restserver"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	v1Group := g.Group("/v1")
	uGroup := v1Group.Group("/user")

	// 需要userClient
	userData := userData.NewUsers()
	// 需要data和jwtOpts
	uService := userService.NewUserService(userData, cfg.Jwt)
	// 需要service
	uController := userController.NewUserController(uService)
	{
		uGroup.POST("/login", uController.Login)
	}
}

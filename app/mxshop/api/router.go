package admin

import (
	"mxshop/app/mxshop/api/config"
	goodsController "mxshop/app/mxshop/api/internal/controller/goods/v1"
	smsController "mxshop/app/mxshop/api/internal/controller/sms/v1"
	userController "mxshop/app/mxshop/api/internal/controller/user/v1"
	userData "mxshop/app/mxshop/api/internal/data/rpc"
	service "mxshop/app/mxshop/api/internal/service"
	smsService "mxshop/app/mxshop/api/internal/service/sms/v1"

	"mxshop/gmicro/server/restserver"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	v1Group := g.Group("/v1")

	// 构造data工厂函数，管理所有data层
	dataFactory, _ := userData.GetDataFactory(cfg.Registry)
	// 构造service工厂函数，管理所有service层
	serviceFactory := service.NewService(dataFactory, cfg.Sms, cfg.Jwt)
	uController := userController.NewUserController(g.Translator(), serviceFactory)
	uGroup := v1Group.Group("/user")
	{
		uGroup.POST("/login", uController.Login)
		uGroup.POST("/register", uController.Register)
		jwtAuth := newJWTAuth(cfg.Jwt)
		uGroup.GET("/detail", jwtAuth.AuthFunc(), uController.GetUserDetail)
		uGroup.PATCH("/update", jwtAuth.AuthFunc(), uController.UpdateUser)
	}
	baseGroup := v1Group.Group("/base")
	{
		smsSrv := smsService.NewSmsService(cfg.Sms)
		smsController.NewSmsController(smsSrv, g.Translator())
		baseGroup.GET("/captcha", userController.GetCaptcha)
		baseGroup.POST("/send_sms")
	}
	gGroup := v1Group.Group("goods")
	gController := goodsController.NewGoodsController(serviceFactory)
	{
		gGroup.GET("/list", gController.List)
	}
}

package admin

import (
	"mxshop/app/mxshop/api/config"
	smsController "mxshop/app/mxshop/api/internal/controller/sms/v1"
	userController "mxshop/app/mxshop/api/internal/controller/user/v1"
	userData "mxshop/app/mxshop/api/internal/data/rpc"
	smsService "mxshop/app/mxshop/api/internal/service/sms/v1"
	userService "mxshop/app/mxshop/api/internal/service/user/v1"
	"mxshop/gmicro/server/restserver"
)

func initRouter(g *restserver.Server, cfg *config.Config) {
	v1Group := g.Group("/v1")

	// 需要userClient
	userData, _ := userData.GetDataFactory(*cfg.Registry)
	// 需要data和jwtOpts
	uService := userService.NewUserService(userData, cfg.Jwt)
	// 需要service
	uController := userController.NewUserController(g.Translator(), uService)
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

}

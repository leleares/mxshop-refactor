package admin

import (
	"mxshop/app/mxshop/admin/controller"
	"mxshop/gmicro/server/restserver"
)

func initRouter(g *restserver.Server) {
	v1 := g.Group("/v1")
	ugroup := v1.Group("/user")
	ucontroller := controller.NewUserController()
	ugroup.GET("list", ucontroller.List)
}

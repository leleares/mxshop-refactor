package user

import (
	"mxshop/app/mxshop/api/internal/service/v1/user"
)

type userServer struct {
	userService user.UserSrv
}

func NewUserController(userService user.UserSrv) *userServer {
	return &userServer{
		userService,
	}
}

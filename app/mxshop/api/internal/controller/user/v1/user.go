package user

import (
	userService "mxshop/app/mxshop/api/internal/service/user/v1"

	ut "github.com/go-playground/universal-translator"
)

type userServer struct {
	trans ut.Translator

	sf userService.UserSrv
}

func NewUserController(trans ut.Translator, sf userService.UserSrv) *userServer {
	return &userServer{trans, sf}
}

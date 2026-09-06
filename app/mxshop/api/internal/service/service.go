package service

import (
	"mxshop/app/mxshop/api/internal/data"
	v14 "mxshop/app/mxshop/api/internal/service/goods/v1"
	v12 "mxshop/app/mxshop/api/internal/service/sms/v1"
	v13 "mxshop/app/mxshop/api/internal/service/user/v1"
	"mxshop/app/pkg/options"
)

type ServiceFactory interface {
	Users() v13.UserSrv
	Sms() v12.SmsSrv
	Goods() v14.GoodsSrv
}

type service struct {
	data data.DataFactory

	smsOpts *options.SmsOptions

	jwtOpts *options.JwtOptions
}

func (s *service) Sms() v12.SmsSrv {
	return v12.NewSmsService(s.smsOpts)
}

func (s *service) Users() v13.UserSrv {
	return v13.NewUserService(s.data, s.jwtOpts)
}

func (s *service) Goods() v14.GoodsSrv {
	return v14.NewGoods(s.data)
}

func NewService(store data.DataFactory, smsOpts *options.SmsOptions, jwtOpts *options.JwtOptions) *service {
	return &service{data: store,
		smsOpts: smsOpts,
		jwtOpts: jwtOpts,
	}
}

var _ ServiceFactory = &service{}

package service

import (
	v1 "mxshop/app/order/srv/internal/data/v1"
	"mxshop/app/pkg/options"
)

type ServiceFactory interface {
	Orders() OrderSrv
}

type service struct {
	data    v1.DataFactory
	dtmopts *options.DtmOptions
}

func (s *service) Orders() OrderSrv {
	return newOrderService(s)
}

var _ ServiceFactory = &service{}

func NewService(data v1.DataFactory, dtmopts *options.DtmOptions) *service {
	return &service{data: data, dtmopts: dtmopts}
}

package v1

import (
	"mxshop/app/goods/srv/data"
	v1 "mxshop/app/goods/srv/data_search"
)

type ServiceFactory interface {
	Goods() GoodsSrv
}

type service struct {
	data       data.DataFactory
	searchData v1.SearchFactory
}

func NewService(data data.DataFactory, searchData v1.SearchFactory) *service {
	return &service{
		data:       data,
		searchData: searchData,
	}
}

func (s *service) Goods() GoodsSrv {
	return newGoods(s)
}

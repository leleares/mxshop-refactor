package dto

import "mxshop/app/order/srv/internal/domain/do"

type OrderDTO struct {
	do.OrderInfoDO
}

type OrderDTOList struct {
	TotalCount int64       `json:"totalCount,omitempty"`
	Items      []*OrderDTO `json:"data"`
}

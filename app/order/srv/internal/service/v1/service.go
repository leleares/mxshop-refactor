package v1

import (
	dv1 "mxshop/app/order/srv/internal/data/v1"
)

type ServiceFactory interface {
	Order() dv1.OrderStore
	ShopCart() dv1.ShopCartStore
}

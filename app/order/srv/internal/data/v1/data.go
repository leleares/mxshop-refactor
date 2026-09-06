package v1

import (
	proto "mxshop/api/goods/v1"
	proto2 "mxshop/api/inventory/v1"

	"gorm.io/gorm"
)

type DataFactory interface {
	Orders() OrderStore
	ShopCarts() ShopCartStore
	Goods() proto.GoodsClient
	Inventorys() proto2.InventoryClient

	Begin() *gorm.DB
}

package v1

import (
	"gorm.io/gorm"
	proto "mxshop/api/goods/v1"
	proto2 "mxshop/api/inventory/v1"
)

type DataFactory interface {
	Orders() OrderStore
	ShopCarts() ShopCartStore
	Goods() proto.GoodsClient
	Inventorys() proto2.InventoryClient

	Begin() *gorm.DB
}

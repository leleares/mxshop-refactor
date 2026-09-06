package data

import (
	v1 "mxshop/app/inventory/srv/internal/data/v1"

	"gorm.io/gorm"
)

type DataFactory interface {
	Inventorys() v1.InventoryStore
	Begin() *gorm.DB
}

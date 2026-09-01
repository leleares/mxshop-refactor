// 工厂store，可以理解为管理所有store的
package data

import "gorm.io/gorm"

type DataFactory interface {
	Goods() GoodsStore
	Categorys() CategoryStore
	Brands() BrandsStore
	Banners() BannerStore
	CategoryBrands() GoodsCategoryBrandStore
	Begin() *gorm.DB
}

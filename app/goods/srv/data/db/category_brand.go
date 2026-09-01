package db

import (
	"context"

	v1 "mxshop/app/goods/srv/data"
	"mxshop/app/goods/srv/domain/do"
	metav1 "mxshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

type categoryBrands struct {
	db *gorm.DB
}

func newCategoryBrand(mf *mysqlFactory) *categoryBrands {
	cb := &categoryBrands{
		db: mf.db,
	}
	return cb
}

func (cb *categoryBrands) List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.GoodsCategoryBrandList, error) {
	//TODO implement me
	panic("implement me")
}

func (cb *categoryBrands) Create(ctx context.Context, txn *gorm.DB, gcb *do.GoodsCategoryBrandDO) error {
	//TODO implement me
	panic("implement me")
}

func (cb *categoryBrands) Update(ctx context.Context, txn *gorm.DB, gcb *do.GoodsCategoryBrandDO) error {
	//TODO implement me
	panic("implement me")
}

func (cb *categoryBrands) Delete(ctx context.Context, ID uint64) error {
	//TODO implement me
	panic("implement me")
}

var _ v1.GoodsCategoryBrandStore = &categoryBrands{}

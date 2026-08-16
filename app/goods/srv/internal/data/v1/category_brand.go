package v1

import (
	"context"
	"mxshop/app/goods/srv/internal/domain/do"

	"gorm.io/gorm"
	metav1 "mxshop/pkg/common/meta/v1"
)

type GoodsCategoryBrandStore interface {
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.GoodsCategoryBrandList, error)
	Create(ctx context.Context, txn *gorm.DB, gcb *do.GoodsCategoryBrandDO) error
	Update(ctx context.Context, txn *gorm.DB, gcb *do.GoodsCategoryBrandDO) error
	Delete(ctx context.Context, ID uint64) error
}

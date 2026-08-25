package data

import (
	"context"
	"mxshop/app/goods/srv/domain/do"
	metav1 "mxshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

type BannerStore interface {
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.BannerList, error)
	Create(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error
	Update(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error
	Delete(ctx context.Context, ID uint64) error
}

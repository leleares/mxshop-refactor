package v1

import (
	"context"
	"gorm.io/gorm"
	"mxshop/app/goods/srv/internal/domain/do"
	metav1 "mxshop/pkg/common/meta/v1"
)

type BannerStore interface {
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.BannerList, error)
	Create(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error
	Update(ctx context.Context, txn *gorm.DB, banner *do.BannerDO) error
	Delete(ctx context.Context, ID uint64) error
}

package v1

import (
	"context"
	"mxshop/app/goods/srv/internal/domain/do"

	"gorm.io/gorm"
	metav1 "mxshop/pkg/common/meta/v1"
)

type BrandsStore interface {
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*do.BrandsDOList, error)
	Create(ctx context.Context, txn *gorm.DB, brands *do.BrandsDO) error
	Update(ctx context.Context, txn *gorm.DB, brands *do.BrandsDO) error
	Delete(ctx context.Context, ID uint64) error
	Get(ctx context.Context, ID uint64) (*do.BrandsDO, error)
}

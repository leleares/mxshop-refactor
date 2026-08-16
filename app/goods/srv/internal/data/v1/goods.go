package v1

import (
	"context"
	"gorm.io/gorm"
	"mxshop/app/goods/srv/internal/domain/do"
	metav1 "mxshop/pkg/common/meta/v1"
)

type GoodsStore interface {
	Get(ctx context.Context, ID uint64) (*do.GoodsDO, error)
	ListByIDs(ctx context.Context, ids []uint64, orderby []string) (*do.GoodsDOList, error)
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.GoodsDOList, error)
	Create(ctx context.Context, goods *do.GoodsDO) error
	CreateInTxn(ctx context.Context, txn *gorm.DB, goods *do.GoodsDO) error
	Update(ctx context.Context, goods *do.GoodsDO) error
	UpdateInTxn(ctx context.Context, txn *gorm.DB, goods *do.GoodsDO) error
	Delete(ctx context.Context, ID uint64) error
	DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error

	Begin() *gorm.DB
}

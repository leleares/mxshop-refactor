package v1

import (
	"context"
	"gorm.io/gorm"
	"mxshop/app/order/srv/internal/domain/do"

	metav1 "mxshop/pkg/common/meta/v1"
)

type ShopCartStore interface {
	List(ctx context.Context, userID uint64, checked bool, meta metav1.ListMeta, orderby []string) (*do.ShoppingCartDOList, error)
	Create(ctx context.Context, cartItem *do.ShoppingCartDO) error
	Get(ctx context.Context, userID, goodsID uint64) (*do.ShoppingCartDO, error)
	UpdateNum(ctx context.Context, cartItem *do.ShoppingCartDO) error
	Delete(ctx context.Context, ID uint64) error
	ClearCheck(ctx context.Context, userID uint64) error

	DeleteByGoodsIDs(ctx context.Context, txn *gorm.DB, userID uint64, goodsIDs []int32) error
}

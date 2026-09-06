package v1

import (
	"context"
	"mxshop/app/order/srv/internal/domain/do"
	metav1 "mxshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

type OrderStore interface {
	Get(ctx context.Context, orderSn string) (*do.OrderInfoDO, error)

	List(ctx context.Context, userID uint64, meta metav1.ListMeta, orderby []string) (*do.OrderInfoDOList, error)

	Create(ctx context.Context, txn *gorm.DB, order *do.OrderInfoDO) error

	Update(ctx context.Context, txn *gorm.DB, order *do.OrderInfoDO) error
}

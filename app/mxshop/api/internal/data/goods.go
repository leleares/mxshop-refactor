package data

import (
	"context"
	v1 "mxshop/api/goods/v1"
)

type GoodsData interface {
	GoodsList(ctx context.Context, req *v1.GoodsFilterRequest) (*v1.GoodsListResponse, error)
}

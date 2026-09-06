package rpc

import (
	"context"
	gpbv1 "mxshop/api/goods/v1"
	"mxshop/gmicro/server/rpcserver"
	"mxshop/gmicro/server/rpcserver/clientinterceptors"

	v1 "mxshop/api/goods/v1"
	"mxshop/gmicro/registry"
)

const goodsserviceName = "discovery:///mxshop-goods-srv"

func NewGoodsServiceClient(r registry.Discovery) gpbv1.GoodsClient {
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(goodsserviceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		panic(err)
	}
	c := gpbv1.NewGoodsClient(conn)
	return c
}

type goods struct {
	gc gpbv1.GoodsClient
}

func newGoods(gc gpbv1.GoodsClient) *goods {
	return &goods{
		gc,
	}
}

func (g *goods) GoodsList(ctx context.Context, req *v1.GoodsFilterRequest) (*v1.GoodsListResponse, error) {
	resp, err := g.gc.GoodsList(ctx, req)
	if err != nil {
		return &v1.GoodsListResponse{}, err
	}
	return resp, nil
}

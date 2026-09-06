package db

import (
	"context"

	proto "mxshop/api/inventory/v1"
	"mxshop/app/pkg/options"
	"mxshop/gmicro/server/rpcserver"
	"mxshop/gmicro/server/rpcserver/clientinterceptors"

	"mxshop/gmicro/registry"
)

const ginvserviceName = "discovery:///mxshop-inventory-srv"

func GetInventoryClient(opts *options.RegistryOptions) proto.InventoryClient {
	discovery := NewDiscovery(opts)
	invClient := NewInventoryServiceClient(discovery)
	return invClient
}

func NewInventoryServiceClient(r registry.Discovery) proto.InventoryClient {
	conn, err := rpcserver.DialInsecure(
		context.Background(),
		rpcserver.WithEndpoint(ginvserviceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	if err != nil {
		panic(err)
	}
	c := proto.NewInventoryClient(conn)
	return c
}

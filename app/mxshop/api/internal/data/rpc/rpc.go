package rpc

import (
	"fmt"
	gpbv1 "mxshop/api/goods/v1"
	upbv1 "mxshop/api/user/v1"
	"mxshop/app/pkg/options"
	"mxshop/gmicro/registry"
	"mxshop/gmicro/registry/consul"
	"sync"

	"mxshop/app/mxshop/api/internal/data"
	ud "mxshop/app/mxshop/api/internal/data"

	cosulAPI "github.com/hashicorp/consul/api"
)

type Factory struct {
	uc upbv1.UserClient
	gc gpbv1.GoodsClient
}

func (f *Factory) Goods() ud.GoodsData {
	return newGoods(f.gc)
}

func (f *Factory) Users() ud.UserData {
	return newUsers(f.uc)
}

func NewDiscovery(opts *options.RegistryOptions) registry.Discovery {
	c := cosulAPI.DefaultConfig()
	c.Address = opts.Address
	c.Scheme = opts.Scheme
	cli, err := cosulAPI.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

var (
	dataFactoryOnce sync.Once
	dataFactory     data.DataFactory
)

// 基于服务发现
func GetDataFactory(registry *options.RegistryOptions) (data.DataFactory, error) {
	if registry == nil && dataFactory == nil {
		return nil, fmt.Errorf("failed to get data factory store")
	}

	dataFactoryOnce.Do(func() {
		d := NewDiscovery(registry)
		// 创建userClient
		userClient := NewUserServiceClient(d)
		goodsClient := NewGoodsServiceClient(d)

		dataFactory = &Factory{
			uc: userClient,
			gc: goodsClient,
		}
	})

	return dataFactory, nil
}

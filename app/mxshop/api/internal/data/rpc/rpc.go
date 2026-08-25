package rpc

import (
	"mxshop/app/mxshop/api/internal/data"
	"mxshop/app/pkg/options"
	"mxshop/gmicro/registry"
	"mxshop/gmicro/registry/consul"

	cosulAPI "github.com/hashicorp/consul/api"
)

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

// 基于服务发现
func GetDataFactory(registry options.RegistryOptions) (data.UserData, error) {
	d := NewDiscovery(&registry)
	// 创建userClient
	userClient := NewUserServiceClient(d)

	users := NewUsers(userClient)
	return users, nil
}

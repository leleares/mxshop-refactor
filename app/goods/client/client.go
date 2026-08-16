package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	v1 "mxshop/api/goods/v1"
	"mxshop/gmicro/registry/consul"
	rpc "mxshop/gmicro/server/rpcserver"
	_ "mxshop/gmicro/server/rpcserver/resolver/direct"
	"mxshop/gmicro/server/rpcserver/selector"
	"mxshop/gmicro/server/rpcserver/selector/random"
	"time"
)

func main() {
	//设置全局的负载均衡策略
	selector.SetGlobalSelector(random.NewBuilder())
	rpc.InitBuilder()

	conf := api.DefaultConfig()
	conf.Address = "127.0.0.1:8500"
	conf.Scheme = "http"
	cli, err := api.NewClient(conf)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))

	conn, err := rpc.DialInsecure(context.Background(),
		rpc.WithBalancerName("selector"),
		rpc.WithDiscovery(r),
		rpc.WithClientTimeout(time.Second*5000),
		rpc.WithEndpoint("discovery:///mxshop-goods-srv"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	uc := v1.NewGoodsClient(conn)

	re, err := uc.GoodsList(context.Background(), &v1.GoodsFilterRequest{
		KeyWords: "猕猴桃",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(re)

}

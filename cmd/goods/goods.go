package main

import (
	"math/rand"
	"mxshop/app/goods/srv"
	"os"
	"runtime"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	// 启动微服务，微服务启动时内部会启动grpc服务
	srv.NewApp("goods-server").Run()
}

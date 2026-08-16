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
	srv.NewApp("goods-server").Run()
}

package v1

import (
	"fmt"
	"mxshop/app/inventory/srv/internal/data"
	"mxshop/app/pkg/options"

	goredislib "github.com/go-redis/redis/v8"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

type ServiceFactory interface {
	Inventorys() InventorySrv
}

type service struct {
	data data.DataFactory

	redisOptions *options.RedisOptions
	pool         redsyncredis.Pool
}

func (s *service) Inventorys() InventorySrv {
	return newInventoryService(s)
}

func NewService(store data.DataFactory, redisOptions *options.RedisOptions) ServiceFactory {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: fmt.Sprintf("%s:%d", redisOptions.Host, redisOptions.Port),
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	return &service{data: store, redisOptions: redisOptions, pool: pool}
}

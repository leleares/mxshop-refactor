package es

import (
	v1 "mxshop/app/goods/srv/data_search"
	"mxshop/app/pkg/options"
	"mxshop/pkg/db"
	"mxshop/pkg/errors"
	"sync"

	"github.com/olivere/elastic/v7"
)

var (
	searchFactory v1.SearchFactory
	once          sync.Once
)

type dataSearch struct {
	esClient *elastic.Client
}

func (ds *dataSearch) Goods() v1.GoodsStore {
	return newGoods(ds)
}

func GetSearchFactoryOr(opts *options.EsOptions) (v1.SearchFactory, error) {
	if opts == nil && searchFactory == nil {
		return nil, errors.New("failed to get es client")
	}

	once.Do(func() {
		esOpt := db.EsOptions{
			Host: opts.Host,
			Port: opts.Port,
		}
		esClient, err := db.NewEsClient(&esOpt)
		if err != nil {
			return
		}
		searchFactory = &dataSearch{esClient: esClient}
	})
	if searchFactory == nil {
		return nil, errors.New("failed to get es client")
	}
	return searchFactory, nil
}

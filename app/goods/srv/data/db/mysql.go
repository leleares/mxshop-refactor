package db

import (
	"fmt"
	"log"
	v1 "mxshop/app/goods/srv/data"
	"mxshop/app/pkg/code"
	"mxshop/app/pkg/options"
	errors2 "mxshop/pkg/errors"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbFactory v1.DataFactory
	once      sync.Once
)

type mysqlFactory struct {
	db *gorm.DB
}

func (mf *mysqlFactory) Begin() *gorm.DB {
	return mf.db.Begin()
}

func (mf *mysqlFactory) Goods() v1.GoodsStore {
	return newGoods(mf)
}

func (mf *mysqlFactory) Categorys() v1.CategoryStore {
	return newCategorys(mf)
}

func (mf *mysqlFactory) Brands() v1.BrandsStore {
	return newBrands(mf)
}

func (mf *mysqlFactory) Banners() v1.BannerStore {
	return newBanner(mf)
}

func (mf *mysqlFactory) CategoryBrands() v1.GoodsCategoryBrandStore {
	return newCategoryBrand(mf)
}

// 这个方法会返回gorm连接
// 这里一定是单例模式
// 这个方法应该返回的是全局的一个变量，如果一开始的时候没有初始化好，那么就初始化一次，后续呢直接拿到这个变量
func GetDBFactoryOr(mysqlOpts *options.MySQLOptions) (v1.DataFactory, error) {
	if mysqlOpts == nil && dbFactory == nil {
		return nil, fmt.Errorf("failed to get mysql store fatory")
	}

	var err error
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			mysqlOpts.Username,
			mysqlOpts.Password,
			mysqlOpts.Host,
			mysqlOpts.Port,
			mysqlOpts.Database)

		// gorm 打印日志，可以自己实现，本质就是一个interface
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,       // Disable color
			},
		)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			return
		}

		sqlDB, _ := db.DB()
		dbFactory = &mysqlFactory{
			db: db,
		}

		sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(mysqlOpts.MaxConnectionLifetime)
	})

	if dbFactory == nil || err != nil {
		return nil, errors2.WithCode(code.ErrConnectDB, "failed to get mysql store factory")
	}
	return dbFactory, nil
}

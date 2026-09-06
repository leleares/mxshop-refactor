package mysql

import (
	"fmt"
	"mxshop/app/inventory/srv/internal/data"
	v1 "mxshop/app/inventory/srv/internal/data/v1"
	"mxshop/app/pkg/options"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlStore struct {
	db *gorm.DB
}

func newMySqlStore(db *gorm.DB) data.DataFactory {
	return &mysqlStore{
		db,
	}
}

func (f *mysqlStore) Inventorys() v1.InventoryStore {
	return newInventorys(f)
}

func (f *mysqlStore) Begin() *gorm.DB {
	return f.db.Begin()
}

var (
	once               sync.Once
	mySqlStoreInstance data.DataFactory
)

func GetFactoryDBOr(mysqlOpts *options.MySQLOptions) (data.DataFactory, error) {
	if mysqlOpts == nil && mySqlStoreInstance == nil {
		return nil, fmt.Errorf("failed to get mysql store fatory")
	}
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			mysqlOpts.Username,
			mysqlOpts.Password,
			mysqlOpts.Host,
			mysqlOpts.Port,
			mysqlOpts.Database)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		sqlDB, _ := db.DB()
		mySqlStoreInstance = newMySqlStore(db)

		sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConnections)
		sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConnections)
		sqlDB.SetConnMaxLifetime(mysqlOpts.MaxConnectionLifetime)
	})

	return mySqlStoreInstance, nil
}

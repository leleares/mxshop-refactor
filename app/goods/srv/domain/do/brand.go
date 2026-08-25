package do

import (
	"context"
	bgorm "mxshop/app/pkg/gorm"
	metav1 "mxshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

type BrandsDO struct {
	bgorm.BaseModel

	Name string `gorm:"type:varchar(20);not null"`
	Logo string `gorm:"type:varchar(200);default:'';not null"`
}

func (BrandsDO) TableName() string {
	return "brands"
}

type BrandsDOList struct {
	TotalCount int64       `json:"totalCount,omitempty"`
	Items      []*BrandsDO `json:"items"`
}

type BrandsStore interface {
	List(ctx context.Context, opts metav1.ListMeta, orderby []string) (BrandsDOList, error)
	Create(ctx context.Context, txn *gorm.DB, brands BrandsDO) error
	Update(ctx context.Context, txn *gorm.DB, brands BrandsDO) error
	Delete(ctx context.Context, ID uint64) error
	Get(ctx context.Context, ID uint64) (BrandsDO, error)
}

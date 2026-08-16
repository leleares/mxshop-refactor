package data

import (
	"context"
	"time"

	metav1 "mxshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

// 基础字段信息（基础表结构信息）
type BaseModel struct {
	ID        int32     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"column:add_time"`
	UpdatedAt time.Time `gorm:"column:update_time"`
	DeletedAt gorm.DeletedAt
	IsDeleted bool
}

// 用户表
type UserDO struct {
	BaseModel
	Mobile   string     `gorm:"type:varchar(11);index:idx_mobile;unique;not null"`
	Password string     `gorm:"type:varchar(100);not null;"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"`
	Gender   string     `gorm:"column:gender;type:varchar(6);default:'male';comment:'female表示女，male表示男'"`
	Role     int        `gorm:"column:role;default:1;type:int;comment:'1表示普通用户，2表示管理员'"`
}

type UserDOList struct {
	Items []*UserDO
	Total int64
}

func (UserDO) TableName() string {
	return "user"
}

type UserStore interface {
	// 注意，有数据访问的操作，一定要有 context 参数，方便做链路追踪，和 error 状态
	GetUserList(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDOList, error)
	GetUserByMobile(ctx context.Context, mobile string) (*UserDO, error)
	GetUserById(ctx context.Context, id int32) (*UserDO, error)
	CreateUser(ctx context.Context, user *UserDO) error
	UpdateUser(ctx context.Context, user *UserDO) error
	DeleteUser(ctx context.Context, id int32) error
}

type userStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) UserStore {
	return &userStore{db: db}
}

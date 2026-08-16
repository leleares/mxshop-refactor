package data

import (
	"context"
	metav1 "mxshop/pkg/common/meta/v1"
)

func (u *userStore) GetUserList(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDOList, error) {
	//实现gorm查询
	ret := UserDOList{}

	//分页
	var limit, offset int
	if opts.PageSize == 0 {
		limit = 10
	} else {
		limit = opts.PageSize
	}

	if opts.Page > 0 {
		offset = (opts.Page - 1) * limit
	}

	//排序
	query := u.db // 这里 db 必须复制出来局部变量，否则影响全局DB
	for _, value := range orderby {
		//坑
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.Total)
	if d.Error != nil {
		return nil, d.Error
	}
	return &ret, nil
}

func (u *userStore) GetUserByMobile(ctx context.Context, mobile string) (*UserDO, error) {
	var user UserDO
	if err := u.db.Where("mobile = ?", mobile).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userStore) GetUserById(ctx context.Context, id int32) (*UserDO, error) {
	var user UserDO
	if err := u.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *userStore) CreateUser(ctx context.Context, user *UserDO) error {
	return u.db.Create(user).Error
}

func (u *userStore) UpdateUser(ctx context.Context, user *UserDO) error {
	return u.db.Save(user).Error
}

func (u *userStore) DeleteUser(ctx context.Context, id int32) error {
	return u.db.Delete(&UserDO{}, id).Error
}

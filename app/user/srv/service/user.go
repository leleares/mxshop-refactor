package service

import (
	"context"
	"mxshop/app/user/srv/data"
	metav1 "mxshop/pkg/common/meta/v1"
)

type UserDTO struct {
	data.UserDO
}

type UserDTOList struct {
	Items []UserDTO
	Total int64
}

type UserServiceInterface interface {
	GetUserList(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDTOList, error)
	GetUserByMobile(ctx context.Context, mobile string) (*UserDTO, error)
	GetUserById(ctx context.Context, id int32) (*UserDTO, error)
	CreateUser(ctx context.Context, user *UserDTO) error
	UpdateUser(ctx context.Context, user *UserDTO) error
	DeleteUser(ctx context.Context, id int32) error
}

type UserService struct {
	userStore data.UserStore
}

func NewUserService(userStore data.UserStore) *UserService {
	return &UserService{
		userStore: userStore,
	}
}

func (s *UserService) GetUserList(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDTOList, error) {
	userList, err := s.userStore.GetUserList(ctx, orderby, opts)
	if err != nil {
		return nil, err
	}

	dtoList := &UserDTOList{
		Items: make([]UserDTO, len(userList.Items)),
		Total: userList.Total,
	}

	for i, user := range userList.Items {
		dtoList.Items[i] = UserDTO{*user}
	}

	return dtoList, nil
}

func (s *UserService) GetUserByMobile(ctx context.Context, mobile string) (*UserDTO, error) {
	user, err := s.userStore.GetUserByMobile(ctx, mobile)
	if err != nil {
		return nil, err
	}
	return &UserDTO{*user}, nil
}

func (s *UserService) GetUserById(ctx context.Context, id int32) (*UserDTO, error) {
	user, err := s.userStore.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserDTO{*user}, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *UserDTO) error {
	userDO := &data.UserDO{
		BaseModel: user.BaseModel,
		Mobile:    user.Mobile,
		Password:  user.Password,
		NickName:  user.NickName,
		Birthday:  user.Birthday,
		Gender:    user.Gender,
		Role:      user.Role,
	}
	return s.userStore.CreateUser(ctx, userDO)
}

func (s *UserService) UpdateUser(ctx context.Context, user *UserDTO) error {
	userDO := &data.UserDO{
		BaseModel: user.BaseModel,
		Mobile:    user.Mobile,
		Password:  user.Password,
		NickName:  user.NickName,
		Birthday:  user.Birthday,
		Gender:    user.Gender,
		Role:      user.Role,
	}
	return s.userStore.UpdateUser(ctx, userDO)
}

func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	return s.userStore.DeleteUser(ctx, id)
}

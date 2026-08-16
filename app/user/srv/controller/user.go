package controller

import (
	"context"
	"crypto/sha512"
	v1 "mxshop/api/user/v1"
	"mxshop/app/user/srv/service"
	"mxshop/pkg/log"
	"strings"
	"time"

	metav1 "mxshop/pkg/common/meta/v1"

	data "mxshop/app/user/srv/data"

	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/protobuf/types/known/emptypb"
)

type userServer struct {
	// grpc 要求
	v1.UnimplementedUserServer
	userService service.UserServiceInterface
}

func (us *userServer) CheckPassWord(ctx context.Context, info *v1.PasswordCheckInfo) (*v1.CheckResponse, error) {
	//校验密码
	options := &password.Options{16, 100, 32, sha512.New}
	passwordInfo := strings.Split(info.EncryptedPassword, "$")
	check := password.Verify(info.Password, passwordInfo[2], passwordInfo[3], options)
	return &v1.CheckResponse{Success: check}, nil
}

func (u *userServer) CreateUser(ctx context.Context, request *v1.CreateUserInfo) (*v1.UserInfoResponse, error) {
	log.Infof("create user function called.")

	userDO := data.UserDO{
		Mobile:   request.Mobile,
		NickName: request.NickName,
		Password: request.PassWord,
	}
	userDTO := service.UserDTO{userDO}

	err := u.userService.CreateUser(ctx, &userDTO)
	if err != nil {
		log.Errorf("create user: %v, error: %v", userDTO, err)
		return nil, err
	}

	userInfoRsp := DTOToResponse(&userDTO)
	return userInfoRsp, nil
}

func (u *userServer) GetUserById(ctx context.Context, request *v1.IdRequest) (*v1.UserInfoResponse, error) {
	log.Infof("get user by id function called.")
	user, err := u.userService.GetUserById(ctx, request.Id)
	if err != nil {
		log.Errorf("get user by id: %s, error: %v", request.Id, err)
		return nil, err
	}

	userInfoRsp := DTOToResponse(user)
	return userInfoRsp, nil
}

func (u *userServer) GetUserByMobile(ctx context.Context, request *v1.MobileRequest) (*v1.UserInfoResponse, error) {
	log.Infof("get user by mobile function called.")
	user, err := u.userService.GetUserByMobile(ctx, request.Mobile)
	if err != nil {
		log.Errorf("get user by mobile: %s, error: %v", request.Mobile, err)
		return nil, err
	}

	userInfoRsp := DTOToResponse(user)
	return userInfoRsp, nil
}

func (u *userServer) GetUserList(ctx context.Context, info *v1.PageInfo) (*v1.UserListResponse, error) {
	log.Info("GetUserList is called")
	srvOpts := metav1.ListMeta{
		Page:     int(info.Pn),
		PageSize: int(info.PSize),
	}
	dtoList, err := u.userService.GetUserList(ctx, []string{}, srvOpts)
	if err != nil {
		return nil, err
	}

	var rsp v1.UserListResponse
	for _, value := range dtoList.Items {
		userRsp := DTOToResponse(&value)
		rsp.Data = append(rsp.Data, userRsp)
	}
	return &rsp, nil
}

func (u *userServer) UpdateUser(ctx context.Context, request *v1.UpdateUserInfo) (*emptypb.Empty, error) {
	log.Infof("update user function called.")

	birthDay := time.Unix(int64(request.BirthDay), 0)
	userDO := data.UserDO{
		BaseModel: data.BaseModel{
			ID: request.Id,
		},
		NickName: request.NickName,
		Gender:   request.Gender,
		Birthday: &birthDay,
	}
	userDTO := service.UserDTO{userDO}

	err := u.userService.UpdateUser(ctx, &userDTO)
	if err != nil {
		log.Errorf("update user: %v, error: %v", userDTO, err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

var _ v1.UserServer = &userServer{}

func DTOToResponse(userDTO *service.UserDTO) *v1.UserInfoResponse {
	userInfoRsp := v1.UserInfoResponse{
		Id:       userDTO.ID,
		PassWord: userDTO.Password,
		NickName: userDTO.NickName,
		Gender:   userDTO.Gender,
		Role:     int32(userDTO.Role),
		Mobile:   userDTO.Mobile,
	}
	if userDTO.Birthday != nil {
		userInfoRsp.BirthDay = uint64(userDTO.Birthday.Unix())
	}
	return &userInfoRsp
}

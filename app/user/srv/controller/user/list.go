package user

import (
	"context"
	upbv1 "mxshop/api/user/v1"
	srvv1 "mxshop/app/user/srv/service/v1"
	metav1 "mxshop/pkg/common/meta/v1"
	"mxshop/pkg/log"
)

func DTOToResponse(userDTO srvv1.UserDTO) *upbv1.UserInfoResponse {
	//在grpc的message中字段有默认值，你不能随便赋值nil进去，容易出错
	//这里要搞清， 哪些字段是有默认值
	userInfoRsp := upbv1.UserInfoResponse{
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

/*
controller 层依赖了service， service层依赖了data层：
controller层能否直接依赖data层： 可以的
controller依赖service并不是直接依赖了具体的struct而是依赖了interface
*/
func (us *userServer) GetUserList(ctx context.Context, info *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	log.Info("GetUserList is called")
	srvOpts := metav1.ListMeta{
		Page:     int(info.Pn),
		PageSize: int(info.PSize),
	}
	dtoList, err := us.srv.List(ctx, []string{}, srvOpts)
	if err != nil {
		return nil, err
	}

	var rsp upbv1.UserListResponse
	for _, value := range dtoList.Items {
		userRsp := DTOToResponse(*value)
		rsp.Data = append(rsp.Data, userRsp)
	}
	return &rsp, nil
}

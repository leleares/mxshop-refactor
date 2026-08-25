package user

import (
	"mxshop/gmicro/server/restserver/middlewares"
	"mxshop/pkg/common/core"

	"github.com/gin-gonic/gin"
)

func (us *userServer) GetUserDetail(ctx *gin.Context) {
	userID, _ := ctx.Get(middlewares.KeyUserID)
	userDTO, err := us.sf.Get(ctx, uint64(userID.(float64)))
	if err != nil {
		core.WriteResponse(ctx, err, nil)
		return
	}
	core.WriteResponse(ctx, nil, gin.H{
		"name":     userDTO.NickName,
		"birthday": userDTO.Birthday.Format("2006-01-02"),
		"gender":   userDTO.Gender,
		"mobile":   userDTO.Mobile,
	})
}

package user

import (
	"mxshop/pkg/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordLoginForm struct {
	Mobile    string `form:"mobile" json:"mobile" binding:"required,mobile"` // 手机号应该是符合某种规范的，怎么进行校验？使用自定义 validator
	Password  string `form:"password" json:"password" binding:"required,min=5,max=20"`
	Captcha   string `form:"captcha" json:"captcha" binding:"required,min=4,max=4"` // 验证码
	CaptchaId string `form:"captcha_id" json:"captcha_id" bindind:"required"`       // 验证码id
}

func (us *userServer) Login(ctx *gin.Context) {
	log.Info("login is called")

	//表单验证
	passwordLoginForm := PasswordLoginForm{}
	if err := ctx.ShouldBind(&passwordLoginForm); err != nil {
		return
	}

	//验证码验证
	if !store.Verify(passwordLoginForm.CaptchaId, passwordLoginForm.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}

	userDTO, err := us.userService.MobileLogin(ctx, passwordLoginForm.Mobile, passwordLoginForm.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "登录失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":        userDTO.ID,
		"nick_name": userDTO.NickName,
		"token":     userDTO.Token,
		// "expired_at": userDTO.ExpiresAt,
	})
}

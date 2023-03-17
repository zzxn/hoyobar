package handler

import (
	"fmt"
	"hoyobar/service"
	"hoyobar/util/myerr"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *service.UserService
}

func (u *UserHandler) AddRoute(r *gin.RouterGroup) {
	r.POST("/test", gin.HandlerFunc(u.CheckOnline))
	r.POST("/verify", gin.HandlerFunc(u.VerifyAccount))
	r.POST("/register", gin.HandlerFunc(u.Register))
	r.POST("/login", gin.HandlerFunc(u.Login))
}

func (u *UserHandler) CheckOnline(c *gin.Context) {
	value := c.Value("user_id")
	if value == nil {
		c.Error(myerr.ErrNotLogin)
		return
	}
	c.JSON(http.StatusOK, fmt.Sprintf("you are online: %T(%v)", value, value))
}

func (u *UserHandler) VerifyAccount(c *gin.Context) {
	req := &AccountVerifyReq{}
	if failBindJSON(c, req) {
		return
	}

	if err := u.UserService.Verify(req.Username); err != nil {
		c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"ecode": "0",
		"emsg":  "已成功发送验证码到邮箱/手机",
	})
}

func (u *UserHandler) Register(c *gin.Context) {
	req := &UserRegisterReq{}
	if failBindJSON(c, req) {
		return
	}

	userBasic, err := u.UserService.Register(&service.RegisterInfo{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Vcode:    req.Vcode,
	})

	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"auth_token": userBasic.AuthToken,
		"username":   req.Username,
		"nickname":   userBasic.Nickname,
		"user_id":    strconv.FormatInt(userBasic.UserID, 10),
	})
	return
}

func (u *UserHandler) Login(c *gin.Context) {
	req := &UserLoginReq{}
	if failBindJSON(c, req) {
		return
	}
	userBasic, err := u.UserService.Login(req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"auth_token": userBasic.AuthToken,
		"username":   req.Username,
		"nickname":   userBasic.Nickname,
		"user_id":    strconv.FormatInt(userBasic.UserID, 10),
	})
	return
}

// func (u *UserHandler) GetUserInfo(c *gin.Context) {
//     userID := c.Query("user_id")
//     if userID == "" {
//         // TODO
//     }
// }

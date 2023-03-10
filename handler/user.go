package handler

import (
	"fmt"
	"hoyobar/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
    UserService *service.UserService
}

func (u *UserHandler) AddRoute(r *gin.RouterGroup) {
    r.POST("/hello", gin.HandlerFunc(u.Hello))
    r.POST("/verify", gin.HandlerFunc(u.VerifyAccount))
    r.POST("/register", gin.HandlerFunc(u.Register))
    r.POST("/login", gin.HandlerFunc(u.Login))
    r.GET("/info", gin.HandlerFunc(u.GetUserInfo))
}

func (u *UserHandler) Hello(c *gin.Context) {
    req := UserHelloReq{}
    if err := c.ShouldBindJSON(&req); err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("failed to parse request body: %v", err))
        return
    } 

    c.JSON(http.StatusOK, fmt.Sprintf("Got your message: %q", req.Msg)) 
}

func (u *UserHandler) VerifyAccount(c *gin.Context) {
   //TODO
}

func (u *UserHandler) Register(c *gin.Context) {
    req := UserRegisterReq{}
    if err := c.ShouldBindJSON(&req); err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("failed to parse request body: %v", err))
        return
    }
    userBasic, err := u.UserService.Register(req.Username, req.Password, req.Vcode)  
    if err != nil {
        c.JSON(http.StatusOK, fmt.Sprintf("err: %v", err))
        return
    }
    c.JSON(http.StatusOK, userBasic)
}

func (u *UserHandler) Login(c *gin.Context) {
    //TODO
}

func (u *UserHandler) GetUserInfo(c *gin.Context) {
    userID := c.Query("user_id")
    if userID == "" {
        // TODO
    }
}


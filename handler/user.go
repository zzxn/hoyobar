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
}

func (u *UserHandler) Hello(c *gin.Context) {
    req := UserHelloReq{}
    if err := c.BindJSON(&req); err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("failed to parse request body: %v", err))
        return
    } 

    c.JSON(http.StatusOK, fmt.Sprintf("Got your message: %q", req.Msg)) 
}

func (u *UserHandler) Register(c *gin.Context) {

}


package handler

import (
	"fmt"
	"hoyobar/util/myerr"

	"github.com/gin-gonic/gin"
	vlid "github.com/go-playground/validator/v10"
)

var validator *vlid.Validate = vlid.New() // thread-safe

func failBindJSON(c *gin.Context, req interface{}) bool {
	// bind req
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(myerr.ErrFailBindJSON.Wrap(err))
		return true
	}

	// validate req
	fmt.Printf("req: %#v\n", req)
	err := validator.Struct(req)
	if err != nil {
		c.Error(myerr.ErrFailBindJSON.Wrap(err))
		return true
	}
	return false
}

type UserRegisterReq struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
	Vcode    string `validate:"required"`
}

type UserLoginReq struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
}

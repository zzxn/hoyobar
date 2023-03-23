package handler

import (
	"context"
	"hoyobar/conf"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	AddRoute(r *gin.RouterGroup)
}

func makeHandlerFunc(f func(context.Context, *gin.Context)) gin.HandlerFunc {
	if conf.Global == nil {
		panic("config is not inited")
	}
	timeout := conf.Global.App.Timeout.Default
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		f(ctx, c)
	})
}

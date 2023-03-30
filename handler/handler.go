package handler

import (
	"github.com/gin-gonic/gin"
)

type Handler interface {
	AddRoute(r *gin.RouterGroup)
}

package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler interface {
	AddRoute(r *gin.RouterGroup)
}

func failBindJSON(c *gin.Context, req interface{}) bool {
    if err := c.ShouldBindJSON(req); err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("failed to parse request body: %v", err))
        return false
    }
    return true
}


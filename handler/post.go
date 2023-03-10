package handler

import (
	"hoyobar/service"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
   PostService *service.PostService 
}

func (p *PostHandler) AddRoute(r *gin.RouterGroup) {

}



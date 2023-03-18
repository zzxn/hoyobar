package handler

import (
	"hoyobar/conf"
	"hoyobar/service"
	"hoyobar/util/myerr"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	PostService *service.PostService
}

func (p *PostHandler) AddRoute(r *gin.RouterGroup) {
	r.POST("/create", gin.HandlerFunc(p.Create))
	r.GET("/detail", gin.HandlerFunc(p.Detail))
	r.GET("/list", gin.HandlerFunc(p.List))
}

func (p *PostHandler) Create(c *gin.Context) {
	req := &PostCreateReq{}
	if failBindJSON(c, req) {
		return
	}
	postID, err := p.PostService.Create(req.AuthorID, req.Title, req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"post_id": strconv.FormatInt(postID, 10),
	})
}

func (p *PostHandler) Detail(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的帖子ID"))
		return
	}
	detail, err := p.PostService.Detail(postID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (p *PostHandler) List(c *gin.Context) {
	var err error
	order := c.Query("order")
	if order == "" {
		order = "create_time"
	}
	cursor := c.Query("cursor")
	pageSizeStr := c.Query("page_size")
	var pageSize int
	if pageSizeStr == "" {
		pageSize = conf.Global.App.DefaultPageSize
	} else if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的页大小"))
	}
	list, err := p.PostService.List(order, cursor, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, list)
}

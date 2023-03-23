package handler

import (
	"context"
	"hoyobar/conf"
	"hoyobar/service"
	"hoyobar/util/myerr"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	PostService *service.PostService
	UserService *service.UserService
}

func (p *PostHandler) AddRoute(r *gin.RouterGroup) {
	r.POST("/create", makeHandlerFunc(p.Create))
	r.POST("/reply", makeHandlerFunc(p.Reply))
	r.GET("/detail", makeHandlerFunc(p.Detail))
	r.GET("/list", makeHandlerFunc(p.List))
	r.GET("/reply/list", makeHandlerFunc(p.ListReply))
}

func (p *PostHandler) userID(c *gin.Context) int64 {
	return c.GetInt64("user_id")
}

func (p *PostHandler) Create(ctx context.Context, c *gin.Context) {
	req := &PostCreateReq{}
	if failBindJSON(c, req) {
		return
	}
	userID := p.userID(c)
	if userID == 0 {
		c.Error(myerr.ErrNotLogin)
		return
	}
	if conf.Global.App.CheckUserIsAuthor && userID != req.AuthorID {
		c.Error(myerr.ErrAuth.WithEmsg("无操作权限"))
	}

	postID, err := p.PostService.Create(ctx, req.AuthorID, req.Title, req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"post_id": strconv.FormatInt(postID, 10),
	})
}

func (p *PostHandler) Reply(ctx context.Context, c *gin.Context) {
	req := &PostReplyReq{}
	if failBindJSON(c, req) {
		return
	}
	userID := p.userID(c)
	if userID == 0 {
		c.Error(myerr.ErrNotLogin)
		return
	}
	if conf.Global.App.CheckUserIsAuthor && userID != req.AuthorID {
		c.Error(myerr.ErrAuth.WithEmsg("无操作权限"))
	}

	replyID, err := p.PostService.Reply(ctx, req.AuthorID, req.PostID, req.Content)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reply_id": strconv.FormatInt(replyID, 10),
	})
}

func (p *PostHandler) Detail(ctx context.Context, c *gin.Context) {
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的帖子ID"))
		return
	}
	detail, err := p.PostService.Detail(ctx, postID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (p *PostHandler) List(ctx context.Context, c *gin.Context) {
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
	list, err := p.PostService.List(ctx, order, cursor, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (p *PostHandler) ListReply(ctx context.Context, c *gin.Context) {
	var err error
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		c.Error(myerr.ErrBadReqBody.WithCause(err).WithEmsg("不合法的帖子"))
		return
	}
	cursor := c.Query("cursor")
	pageSizeStr := c.Query("page_size")
	var pageSize int
	if pageSizeStr == "" {
		pageSize = conf.Global.App.DefaultPageSize
	} else if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的页大小"))
	}
	list, err := p.PostService.ListReply(ctx, postID, cursor, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, list)
}

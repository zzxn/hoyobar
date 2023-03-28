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
	UserService *service.UserService
}

func (p *PostHandler) AddRoute(r *gin.RouterGroup) {
	r.POST("/create", gin.HandlerFunc(p.Create))
	r.POST("/reply", gin.HandlerFunc(p.Reply))
	r.GET("/detail", gin.HandlerFunc(p.Detail))
	r.GET("/list", gin.HandlerFunc(p.List))
	r.GET("/reply/list", gin.HandlerFunc(p.ListReply))
}

func (p *PostHandler) userID(c *gin.Context) int64 {
	return c.GetInt64("user_id")
}

func (p *PostHandler) Create(c *gin.Context) {
	req := &PostCreateReq{}
	if failBindJSON(c, req) {
		return
	}
	userID := p.userID(c)
	if userID == 0 {
		c.Error(myerr.ErrNotLogin) // nolint:errcheck
		return
	}
	if conf.Global.App.CheckUserIsAuthor && userID != req.AuthorID {
		c.Error(myerr.ErrAuth.WithEmsg("无操作权限")) // nolint:errcheck
	}

	postID, err := p.PostService.Create(c, req.AuthorID, req.Title, req.Content)
	if err != nil {
		c.Error(err) // nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"post_id": strconv.FormatInt(postID, 10),
	})
}

func (p *PostHandler) Reply(c *gin.Context) {
	req := &PostReplyReq{}
	if failBindJSON(c, req) {
		return
	}
	userID := p.userID(c)
	if userID == 0 {
		c.Error(myerr.ErrNotLogin) // nolint:errcheck
		return
	}
	if conf.Global.App.CheckUserIsAuthor && userID != req.AuthorID {
		c.Error(myerr.ErrAuth.WithEmsg("无操作权限")) // nolint:errcheck
	}

	replyID, err := p.PostService.Reply(c, req.AuthorID, req.PostID, req.Content)
	if err != nil {
		c.Error(err) // nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reply_id": strconv.FormatInt(replyID, 10),
	})
}

func (p *PostHandler) Detail(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的帖子ID")) // nolint:errcheck
		return
	}
	detail, err := p.PostService.Detail(c, postID)
	if err != nil {
		c.Error(err) // nolint:errcheck
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
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的页大小")) // nolint:errcheck
	}
	list, err := p.PostService.List(c, order, cursor, pageSize)
	if err != nil {
		c.Error(err) // nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, list)
}

func (p *PostHandler) ListReply(c *gin.Context) {
	var err error
	postID, err := strconv.ParseInt(c.Query("post_id"), 10, 64)
	if err != nil {
		c.Error(myerr.ErrBadReqBody.WithCause(err).WithEmsg("不合法的帖子")) // nolint:errcheck
		return
	}
	cursor := c.Query("cursor")
	pageSizeStr := c.Query("page_size")
	var pageSize int
	if pageSizeStr == "" {
		pageSize = conf.Global.App.DefaultPageSize
	} else if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
		c.Error(myerr.ErrBadReqBody.WithEmsg("不合法的页大小")) // nolint:errcheck
	}
	list, err := p.PostService.ListReply(c, postID, cursor, pageSize)
	if err != nil {
		c.Error(err) // nolint:errcheck
		return
	}
	c.JSON(http.StatusOK, list)
}

package service

import (
	"fmt"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/util/funcs"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PostService struct {
	cache mycache.Cache
}

func NewPostService(cache mycache.Cache) *PostService {
	if conf.Global == nil {
		log.Fatalf("conf.Global is not initialized")
	}
	postService := &PostService{
		cache: cache,
	}
	return postService
}

type PostDetail struct {
	PostID      int64     `json:"post_id,string"`
	AuthorID    int64     `json:"author_id,string"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	CreatedTime time.Time `json:"created_at"`
	ReplyTime   time.Time `json:"reply_time"`
	ReplyNum    int64     `json:"reply_num"`
}

type PostList struct {
	List   []PostDetail `json:"list"`
	Cursor string       `json:"cursor"`
}

type ReplyList struct {
	List []struct {
		ReplyID   int64     `json:"reply_id,string"`
		AuthorID  string    `json:"author_id,string"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
	}
	Cursor string
}

func (p *PostService) Create(authorID int64, title string, content string) (postID int64, err error) {
	// TODO: check authentication and authority
	postID = idgen.New()
	postM := model.Post{
		PostID:    postID,
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		ReplyTime: time.Now(),
		ReplyNum:  0,
	}
	err = model.DB.Create(&postM).Error
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to create post data")
	}
	return postID, nil
}

func (p *PostService) Detail(postID int64) (detail *PostDetail, err error) {
	postM := model.Post{}
	err = model.DB.Model(&model.Post{}).Where("post_id = ?", postID).First(&postM).Error
	if err == gorm.ErrRecordNotFound {
		return nil, myerr.ErrResourceNotFound.WithEmsg("帖子不存在")
	}
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to query post %v", postID)
	}
	return &PostDetail{
		PostID:      postID,
		AuthorID:    postM.AuthorID,
		Title:       postM.Title,
		Content:     postM.Content,
		CreatedTime: postM.CreatedAt,
		ReplyTime:   postM.ReplyTime,
		ReplyNum:    postM.ReplyNum,
	}, nil
}

// order: one of "create_time" and "reply_time", desc order
// cursor: the cursor returned by last call with the same params
func (p *PostService) List(order string, cursor string, pageSize int) (list *PostList, err error) {
	if pageSize <= 0 {
		return nil, myerr.ErrBadReqBody.WithEmsg("页为空")
	}
	pageSize = funcs.Min(pageSize, conf.Global.App.MaxPageSize)
	lastID, lastTime, err := decomposePageCursor(cursor)
	if err != nil {
		return nil, myerr.ErrBadReqBody.WithEmsg("页游标错误")
	}
	// TODO: sort according to reply time
	var orderField string
	switch order {
	case "create_time":
		orderField = "created_at"
	case "reply_time":
		orderField = "reply_time"
	default:
		return nil, myerr.ErrBadReqBody.WithEmsg("不支持的排序方式")
	}

	var posts []model.Post
	err = model.DB.Model(&model.Post{}).
		Where(fmt.Sprintf("%v <= ?", orderField), lastTime).
		Where("post_id < ?", lastID).
		Order(fmt.Sprintf("%v DESC", orderField)).
		Order("post_id DESC").
		Limit(5).
		Find(&posts).Error
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to query post")
	}
	if len(posts) == 0 {
		return nil, myerr.ErrNoMoreEntry
	}

	n := len(posts)
	var newCursor string
	switch order {
	case "create_time":
		newCursor = composePageCursor(posts[n-1].PostID, posts[n-1].CreatedAt)
	case "reply_time":
		newCursor = composePageCursor(posts[n-1].PostID, posts[n-1].ReplyTime)
	default:
		return nil, myerr.ErrBadReqBody.WithEmsg("不支持的排序方式-2")
	}

	list = &PostList{Cursor: newCursor}
	for _, post := range posts {
		list.List = append(list.List, PostDetail{
			PostID:      post.PostID,
			AuthorID:    post.AuthorID,
			Title:       post.Title,
			Content:     post.Content,
			CreatedTime: post.CreatedAt,
			ReplyTime:   post.ReplyTime,
			ReplyNum:    post.ReplyNum,
		})
	}
	return list, nil
}

func (p *PostService) Reply(authorID int64, postID int64, content string) (err error) {
	return
}

func (p *PostService) ListReply(postID int64, cursor string) (list *ReplyList, err error) {
	return
}

var pageCursorTimeFormat = time.RFC3339Nano

func decomposePageCursor(cursor string) (ID int64, t time.Time, err error) {
	if cursor == "" {
		return math.MaxInt64, time.Now(), nil
	}
	segs := strings.SplitN(strings.ReplaceAll(cursor, "P", "+"), "_", 2)
	if len(segs) != 2 {
		return ID, t, fmt.Errorf("Wrong page cursor format: %v", cursor)
	}

	ID, err = strconv.ParseInt(segs[0], 10, 64)
	if err != nil {
		return ID, t, fmt.Errorf("wrong page cursor format, expect first part a int, got %v", segs[0])
	}

	t, err = time.Parse(pageCursorTimeFormat, segs[1])
	if err != nil {
		return ID, t, fmt.Errorf(
			"wrong page cursor format, expect second part time formatted, %v, got %v",
			pageCursorTimeFormat,
			segs[1],
		)
	}

	return ID, t, nil
}

func composePageCursor(ID int64, t time.Time) (cursor string) {
	cursor = fmt.Sprintf("%d_%v", ID, t.Format(pageCursorTimeFormat))
	cursor = strings.ReplaceAll(cursor, "+", "P") // make it url safe
	return cursor
}

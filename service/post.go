package service

import (
	"fmt"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	PostID    int64     `json:"post_id,string"`
	AuthorID  int64     `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreateAt  time.Time `json:"create_at"`
	ReplyTime time.Time `json:"reply_time"`
	ReplyNum  int64     `json:"reply_num"`
}

type PostList struct {
	List   []PostDetail `json:"list"`
	Cursor string       `json:"cursor"`
}

type ReplyList struct {
	List []struct {
		ReplyID  int64     `json:"reply_id"`
		AuthorID string    `json:"author_id"`
		Content  string    `json:"content"`
		CreateAt time.Time `json:"create_at"`
	}
	Cursor string
}

const (
	PostListOrderLatestReply = "lastest_reply"
	PostListOrderLatestPost  = "lastest_post"
)

func (p *PostService) Create(authorID int64, title string, content string) (postID int64, err error) {
	// TODO: check authentication and authority
	postID = idgen.New()
	postM := model.Post{
		PostID:   postID,
		AuthorID: authorID,
		Title:    title,
		Content:  content,
	}
	err = model.DB.Transaction(func(tx *gorm.DB) (err error) {
		err = tx.Create(&postM).Error
		if err != nil {
			return errors.Wrap(err, "fail to create post")
		}

		err = tx.Create(&model.PostStat{
			PostID:    postID,
			ReplyTime: time.Now(),
			ReplyNum:  0,
		}).Error
		if err != nil {
			return errors.Wrap(err, "fail to create post_stat")
		}
		return nil
	})
	if err != nil {
		return 0, myerr.NewOtherErr(err, "fail to create post data")
	}
	return postID, nil
}

func (p *PostService) Detail(postID int64) (detail *PostDetail, err error) {
	postM, postStatM := model.Post{}, model.PostStat{}
	err = model.DB.Model(&model.Post{}).Where("post_id = ?", postID).First(&postM).Error
	if err == gorm.ErrRecordNotFound {
		return nil, myerr.ErrResourceNotFound
	}
	if err != nil {
		return nil, myerr.NewOtherErr(err, "fail to queyr post %v", postID)
	}
	// TODO: use cache here
	err = model.DB.Model(&model.PostStat{}).Where("post_id = ?", &postStatM).Error
	if err != nil {
		return nil, myerr.NewOtherErr(err, "fail to queyr post_stat %v", postID)
	}
	return &PostDetail{
		PostID:    postID,
		AuthorID:  postM.AuthorID,
		Title:     postM.Title,
		Content:   postM.Content,
		CreateAt:  postM.CreatedAt,
		ReplyTime: postStatM.ReplyTime,
		ReplyNum:  postStatM.ReplyNum,
	}, nil
}

func (p *PostService) List(order string, cursor string, authorID int64) (list *PostList, err error) {
	return
}

func (p *PostService) Reply(authorID int64, postID int64, content string) (err error) {
	return
}

func (p *PostService) ReplyList(postID int64, cursor string) (list *ReplyList, err error) {
	return
}

// TODO: 确保时间区域的精度足够
func decomposePageCursor(cursor string) (ID int64, t time.Time, err error) {
	segs := strings.SplitN(cursor, "|", 2)
	if len(segs) != 2 {
		return ID, t, fmt.Errorf("Wrong page cursor format: %v", cursor)
	}

	ID, err = strconv.ParseInt(segs[0], 10, 64)
	if err != nil {
		return ID, t, fmt.Errorf("wrong page cursor format, expect first part a int, got %v", segs[0])
	}

	t, err = time.Parse(time.StampNano, segs[1])
	if err != nil {
		return ID, t, fmt.Errorf(
			"wrong page cursor format, expect second part time formatted, %v, got %v",
			time.StampNano,
			segs[1],
		)
	}

	return ID, t, nil
}

func composePageCursor(ID int64, t time.Time) (cursor string) {
	return fmt.Sprintf("%d|%v", ID, t.Format(time.StampNano))
}

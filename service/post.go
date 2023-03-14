package service

import (
	"fmt"
	"hoyobar/conf"
	"hoyobar/util/mycache"
	"log"
	"strconv"
	"strings"
	"time"
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
    PostID int64
    AuthorID int64
    Title string
    Content string
    CreateAt time.Time
    ReplyTime time.Time
    ReplyCount int64
}

type ReplyList struct {
    list []struct{
        ReplyID int64
        AuthorID string
        Content string
        CreateAt time.Time
    }
    cursor string
}

func (p *PostService) Create(authorID int64, title string, content string) (postID int64, err error) {
    return
}

func (p *PostService) Detail(postID int64) (detail *PostDetail, err error) {
    return
}

func (p *PostService) Reply(authorID int64, postID int64, content string) (err error) {
    return
}

func (p *PostService) FetchReplyList(postID int64, cursor string) (list *ReplyList, err error) {
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

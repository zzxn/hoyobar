package service

import (
	"context"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/storage"
	"hoyobar/util/funcs"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"log"
	"strings"
	"time"
)

type PostService struct {
	cache        mycache.Cache
	userStorage  storage.UserStorage
	postStorage  storage.PostStorage
	replyStorage storage.PostReplyStorage
}

func NewPostService(
	cache mycache.Cache,
	userStorage storage.UserStorage,
	postStorage storage.PostStorage,
	replyStorage storage.PostReplyStorage,
) *PostService {
	if conf.Global == nil {
		log.Fatalf("conf.Global is not initialized")
	}
	postService := &PostService{
		cache:        cache,
		userStorage:  userStorage,
		postStorage:  postStorage,
		replyStorage: replyStorage,
	}
	return postService
}

type PostDetail struct {
	PostID         int64     `json:"post_id,string"`
	AuthorID       int64     `json:"author_id,string"`
	AuthorNickname string    `json:"author_nickname"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	CreatedTime    time.Time `json:"created_at"`
	ReplyTime      time.Time `json:"reply_time"`
	ReplyNum       int64     `json:"reply_num"`
}

type PostList struct {
	List   []PostDetail `json:"list"`
	Cursor string       `json:"cursor"`
}

type ReplyDetail struct {
	ReplyID   int64     `json:"reply_id,string"`
	AuthorID  int64     `json:"author_id,string"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type ReplyList struct {
	List   []ReplyDetail `json:"list"`
	Cursor string        `json:"cursor"`
}

func (p *PostService) Create(ctx context.Context, authorID int64, title string, content string) (postID int64, err error) {
	postID = idgen.New()
	postM := model.Post{
		PostID:    postID,
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		ReplyTime: time.Now(),
		ReplyNum:  0,
	}
	err = p.postStorage.Create(ctx, &postM)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to create post data")
	}
	return postID, nil
}

func (p *PostService) Detail(ctx context.Context, postID int64) (detail *PostDetail, err error) {
	postM, err := p.postStorage.FetchByPostID(ctx, postID)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to query post %v", postID)
	}
	if postM == nil {
		return nil, myerr.ErrResourceNotFound.WithEmsg("帖子不存在")
	}
	var authorNickname string
	author, err := p.userStorage.FetchByUserID(ctx, postM.AuthorID)
	if err != nil && author != nil {
		authorNickname = author.Nickname
	}
	return &PostDetail{
		PostID:         postID,
		AuthorID:       postM.AuthorID,
		AuthorNickname: authorNickname,
		Title:          postM.Title,
		Content:        postM.Content,
		CreatedTime:    postM.CreatedAt,
		ReplyTime:      postM.ReplyTime,
		ReplyNum:       postM.ReplyNum,
	}, nil
}

// order: one of "create_time" and "reply_time", desc order
// cursor: the cursor returned by last call with the same params
func (p *PostService) List(ctx context.Context, order string, cursor string, pageSize int) (list *PostList, err error) {
	if pageSize <= 0 {
		return nil, myerr.ErrBadReqBody.WithEmsg("页为空")
	}
	pageSize = funcs.Min(pageSize, conf.Global.App.MaxPageSize)
	postMs, newCursor, err := p.postStorage.List(ctx, order, cursor, pageSize)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to query posts")
	}
	if len(postMs) == 0 {
		return nil, myerr.ErrNoMoreEntry.WithEmsg("没有更多帖子了")
	}
	list = &PostList{Cursor: newCursor}
	for _, post := range postMs {
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

func (p *PostService) Reply(ctx context.Context, authorID int64, postID int64, content string) (replyID int64, err error) {
	// check params
	content = strings.TrimSpace(content)
	if content == "" {
		return 0, myerr.ErrBadReqBody.WithEmsg("内容不能为空")
	}
	userExist, err := p.userStorage.HasUser(ctx, authorID)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to query user %v", authorID)
	}
	if !userExist {
		return 0, myerr.ErrResourceNotFound.WithEmsg("用户不存在")
	}
	postExist, err := p.postStorage.HasPost(ctx, postID)
	if err != nil {
		return 0, err
	}
	if !postExist {
		return 0, myerr.ErrResourceNotFound.WithEmsg("帖子不存在")
	}

	// create reply
	replyM := model.PostReply{
		ReplyID:  idgen.New(),
		AuthorID: authorID,
		PostID:   postID,
		Content:  content,
	}
	err = p.replyStorage.Create(ctx, &replyM)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to create post reply")
	}

	// update post's reply time
	err = p.postStorage.IncrementReplyNum(ctx, postID, 1)
	if err != nil {
		// minor err, log and ignore
		log.Printf("fails to update reply time, post_id = %v\n", postID)
	}

	return replyM.ReplyID, nil
}

func (p *PostService) ListReply(ctx context.Context, postID int64, cursor string, pageSize int) (list *ReplyList, err error) {
	// check params
	if pageSize <= 0 {
		return nil, myerr.ErrBadReqBody.WithEmsg("页为空")
	}
	pageSize = funcs.Min(pageSize, conf.Global.App.MaxPageSize)

	postExist, err := p.postStorage.HasPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	if !postExist {
		return nil, myerr.ErrResourceNotFound.WithEmsg("帖子不存在")
	}

	// find replies
	replies, newCursor, err := p.replyStorage.List(ctx, postID, storage.PostReplyOrderCreateTimeDesc, cursor, pageSize)
	if len(replies) == 0 {
		return nil, myerr.ErrNoMoreEntry
	}
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to query post reply")
	}

	list = &ReplyList{Cursor: newCursor}
	for _, reply := range replies {
		list.List = append(list.List, ReplyDetail{
			ReplyID:   reply.ReplyID,
			AuthorID:  reply.AuthorID,
			Content:   reply.Content,
			CreatedAt: reply.CreatedAt,
		})
	}
	return list, nil
}

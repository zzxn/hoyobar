package service

import (
	"context"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/storage"
	"hoyobar/util/funcs"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/mycache/keys"
	"hoyobar/util/myerr"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
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
	ReplyID        int64     `json:"reply_id,string"`
	AuthorID       int64     `json:"author_id,string"`
	AuthorNickname string    `json:"author_nickname"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type ReplyList struct {
	List   []ReplyDetail `json:"list"`
	Cursor string        `json:"cursor"`
}

func (p *PostService) Create(ctx context.Context, authorID int64, title string, content string) (postID int64, err error) {
	postID = idgen.New()
	now := funcs.NowInMs()
	postM := model.Post{
		PostID:    postID,
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		ReplyTime: now,
		CreatedAt: now,
		Model: model.Model{
			UpdatedAt: now,
		},
		ReplyNum: 0,
	}
	err = p.postStorage.Create(ctx, &postM)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to create post data")
	}

	// insert into cache zset to accelerate pagination
	if tos, ok := p.cache.(mycache.TimeOrderedSetCache); ok {
		name := keys.PostListName(storage.PostOrderCreateTimeDesc)
		maxSize := conf.Global.App.PostPaginationCacheSize
		err := tos.TOSAdd(ctx, name, mycache.TOSItem{
			T:     now,
			Value: funcs.FullLeadingZeroItoa(postID),
		}, maxSize)
		if err != nil {
			log.Printf("fail to insert created post into cache list, err=%v", err)
		}
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

	list, err = p.listByCache(ctx, order, cursor, pageSize)
	if err == nil && len(list.List) >= pageSize {
		// we are done, shortage of number is possible, but it's ok
		return list, nil
	}
	if err == nil && len(list.List) < pageSize {
		log.Printf("did not query enough posts list cache, expect %v, got %v", pageSize, len(list.List))
	}
	if err != nil {
		log.Printf("fail to query post list in cache, err=%v", err)
		err = nil // nolint: ineffassign
		// clear this err
	}

	// we don't get enough items from cache or fail to get items from cache.
	// try regular database.
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
	list.List = p.fillAuthorNickname(ctx, list.List)
	return list, nil
}

// returned error need to be wrapped
func (p *PostService) listByCache(ctx context.Context, order string, cursor string, pageSize int) (list *PostList, err error) {
	tos, ok := p.cache.(mycache.TimeOrderedSetCache)
	if !ok {
		return nil, errors.Errorf("current cache not support TimeOrderedSetCache interface")
	}

	lastID, lastTime, err := storage.DecomposePageCursor(cursor)
	if err != nil {
		return nil, errors.Wrapf(err, "wrong cursor: %v", cursor)
	}

	lastIDStr := funcs.FullLeadingZeroItoa(lastID)

	tosItems, err := tos.TOSFetch(ctx, keys.PostListName(order), lastTime, lastIDStr, pageSize)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to query post list from cache")
	}
	if len(tosItems) == 0 {
		log.Printf("not get tos items from cache with order=%v, cursor=%v\n", order, cursor)
		// return empty list
		return &PostList{
			List:   []PostDetail{},
			Cursor: cursor,
		}, nil
	}

	// extract postIDs, then map them into PostDetail
	postIDs := make([]int64, len(tosItems))
	for i, t := range tosItems {
		postID, err := strconv.ParseInt(t.Value, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to parse tos item value, expect int64, got %v", t.Value)
		}
		postIDs[i] = postID
	}
	postDetails, err := p.mapIDToDetail(ctx, postIDs)
	if err != nil {
		return nil, errors.Wrap(err, "fail to map post id to detail")
	}

	// compose new cursor
	var newCursor string
	n := len(postDetails)
	switch order {
	case storage.PostOrderCreateTimeDesc:
		newCursor = storage.ComposePageCursor(postDetails[n-1].PostID, postDetails[n-1].CreatedTime)
	case storage.PostOrderReplyTimeDesc:
		newCursor = storage.ComposePageCursor(postDetails[n-1].PostID, postDetails[n-1].ReplyTime)
		for i := 0; i < n; i++ {
			postDetails[i].ReplyTime = tosItems[i].T.Local()
		}
	}

	postDetails = p.fillAuthorNickname(ctx, postDetails)
	return &PostList{
		List:   postDetails,
		Cursor: newCursor,
	}, nil
}

// mapIDToDetail maps postIDs to postDetails
//
// All values in list will be non-nil, fail map one, fail all.
// Minor error will be ignored, such as fail map author id to author nickname.
// The order of postIDs will be kept, i.e result[i].PostID == postIDs[i]
func (p *PostService) mapIDToDetail(ctx context.Context, postIDs []int64) ([]PostDetail, error) {
	postMs, err := p.postStorage.BatchFetchByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to query posts")
	}

	list := make([]PostDetail, len(postIDs))
	for i, post := range postMs {
		if post == nil {
			list[i].Title = "[已删除]"
			continue
		}
		list[i] = PostDetail{
			PostID:      post.PostID,
			AuthorID:    post.AuthorID,
			Title:       post.Title,
			Content:     post.Content,
			CreatedTime: post.CreatedAt,
			ReplyTime:   post.ReplyTime,
			ReplyNum:    post.ReplyNum,
		}
	}
	list = p.fillAuthorNickname(ctx, list)
	return list, nil
}

// ignore fails
func (p *PostService) fillAuthorNickname(ctx context.Context, details []PostDetail) []PostDetail {
	authorIDs := make([]int64, 0, len(details))
	for _, pd := range details {
		authorIDs = append(authorIDs, pd.AuthorID)
	}
	authors, err := p.userStorage.BatchFetchByUserIDs(ctx, authorIDs, []string{"nickname"})
	if err != nil {
		log.Printf("fail to fill author nicknames of posts: %v\n", err)
		return details
	}
	for i, author := range authors {
		if author != nil {
			details[i].AuthorNickname = author.Nickname
		}
	}
	return details
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
	replyTime := funcs.NowInMs()
	err = p.postStorage.IncrementReplyNum(ctx, postID, 1, replyTime)
	if err != nil {
		// minor err, log and ignore
		log.Printf("fails to update reply time, post_id = %v\n", postID)
	}

	// insert into cache zset to accelerate pagination
	if tos, ok := p.cache.(mycache.TimeOrderedSetCache); ok {
		name := keys.PostListName(storage.PostOrderReplyTimeDesc)
		maxSize := conf.Global.App.PostPaginationCacheSize
		err := tos.TOSAdd(ctx, name, mycache.TOSItem{
			T:     replyTime,
			Value: funcs.FullLeadingZeroItoa(postID),
		}, maxSize)
		if err != nil {
			log.Printf("fail to insert created post into cache list, err=%v", err)
		}
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
	list.List = p.fillReplyAuthorNickname(ctx, list.List)
	return list, nil
}

// ignore fails
func (p *PostService) fillReplyAuthorNickname(ctx context.Context, details []ReplyDetail) []ReplyDetail {
	authorIDs := make([]int64, 0, len(details))
	for _, pd := range details {
		authorIDs = append(authorIDs, pd.AuthorID)
	}
	authors, err := p.userStorage.BatchFetchByUserIDs(ctx, authorIDs, []string{"nickname"})
	if err != nil {
		log.Printf("fail to fill author nicknames of replies: %v\n", err)
		return details
	}
	for i, author := range authors {
		if author != nil {
			details[i].AuthorNickname = author.Nickname
		}
	}
	return details
}

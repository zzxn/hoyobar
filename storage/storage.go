package storage

import (
	"context"
	"hoyobar/model"
	"time"
)

type UserStorage interface {
	Create(ctx context.Context, user *model.User) error
	FetchByUserID(ctx context.Context, userID int64) (*model.User, error)
	// BatchFetchByUserIDs allows userIDs duplicate and invalid.
	// The coresponding *model.User of invalid ID will be nil.
	BatchFetchByUserIDs(ctx context.Context, userIDs []int64, fields []string) ([]*model.User, error)
	HasUser(ctx context.Context, userID int64) (bool, error)
	PhoneToUserID(ctx context.Context, phone string) (int64, error)
	EmailToUserID(ctx context.Context, email string) (int64, error)
	NicknameToUserID(ctx context.Context, nickname string) (int64, error)
}

const (
	PostOrderCreateTimeDesc = "create_time"
	PostOrderReplyTimeDesc  = "reply_time"
)

const (
	PostReplyOrderCreateTimeDesc = "create_time"
)

type PostStorage interface {
	Create(ctx context.Context, post *model.Post) error
	FetchByPostID(ctx context.Context, postID int64) (*model.Post, error)
	// BatchFetchByPostIDs allows postIDs duplicate and invalid.
	// The coresponding *model.Post of invalid ID will be nil.
	BatchFetchByPostIDs(ctx context.Context, postIDs []int64) ([]*model.Post, error)
	HasPost(ctx context.Context, postID int64) (bool, error)
	List(ctx context.Context, order string, cursor string, cnt int) (list []*model.Post, newCursor string, err error)
	IncrementReplyNum(ctx context.Context, postID int64, incr int, replyTime time.Time) error
}

type PostReplyStorage interface {
	Create(ctx context.Context, reply *model.PostReply) error
	List(ctx context.Context, postID int64, order string, cursor string, cnt int) (list []*model.PostReply, newCursor string, err error)
}

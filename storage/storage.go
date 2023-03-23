package storage

import (
	"hoyobar/model"
)

type UserStorage interface {
	Create(user *model.User) error
	FetchByUserID(userID int64) (*model.User, error)
	HasUser(userID int64) (bool, error)
	PhoneToUserID(phone string) (int64, error)
	EmailToUserID(email string) (int64, error)
	NicknameToUserID(nickname string) (int64, error)
}

const (
	PostOrderCreateTimeDesc = "create_time"
	PostOrderReplyTimeDesc  = "reply_time"
)

const (
	PostReplyOrderCreateTimeDesc = "create_time"
)

type PostStorage interface {
	Create(post *model.Post) error
	FetchByPostID(postID int64) (*model.Post, error)
	HasPost(postID int64) (bool, error)
	List(order string, cursor string, cnt int) (list []*model.Post, newCursor string, err error)
	IncrementReplyNum(postID int64, incr int) error
}

type PostReplyStorage interface {
	Create(reply *model.PostReply) error
	List(postID int64, order string, cursor string, cnt int) (list []*model.PostReply, newCursor string, err error)
}

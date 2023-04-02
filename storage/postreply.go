package storage

import (
	"context"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/util/funcs"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type PostReplyStorageMySQL struct {
	db *gorm.DB
}

var _ = PostReplyStorage(new(PostReplyStorageMySQL))

func NewPostReplyStorageMySQL(db *gorm.DB) *PostReplyStorageMySQL {
	return &PostReplyStorageMySQL{
		db: db,
	}
}

// Create implements PostReplyStorage
func (p *PostReplyStorageMySQL) Create(ctx context.Context, reply *model.PostReply) error {
	err := p.db.Model(reply).Create(reply).Error
	if err != nil {
		return errors.Wrapf(err, "fail to create post reply")
	}
	return nil
}

// List implements PostReplyStorage
func (p *PostReplyStorageMySQL) List(ctx context.Context, postID int64, order string, cursor string, cnt int) (list []*model.PostReply, newCursor string, err error) {
	cnt = funcs.Clip(cnt, 1, conf.Global.App.MaxPageSize)
	lastID, lastTime, err := DecomposePageCursor(cursor)
	if err != nil {
		return nil, "", errors.Wrapf(err, "wrong cursor: %v", cursor)
	}

	if order != PostReplyOrderCreateTimeDesc {
		return nil, "", errors.Errorf("unsupported post list order: %v", order)
	}

	// find replies
	err = p.db.Model(&model.PostReply{}).
		Where("post_id = ?", postID).
		Where("created_at <= ?", lastTime).
		Where("reply_id < ?", lastID).
		Order("created_at DESC").
		Order("post_id DESC").
		Limit(cnt).
		Find(&list).Error
	if err != nil {
		return nil, "", errors.Wrapf(err, "fail to query post reply list")
	}
	if len(list) == 0 {
		return nil, cursor, nil
	}

	n := len(list)
	newCursor = ComposePageCursor(list[n-1].ReplyID, list[n-1].CreatedAt)
	return list, newCursor, nil
}

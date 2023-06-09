package storage

import (
	"context"
	"fmt"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/util/funcs"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type PostStorageMySQL struct {
	db *gorm.DB
}

var _ = PostStorage(new(PostStorageMySQL))

func NewPostStorageMySQL(db *gorm.DB) *PostStorageMySQL {
	return &PostStorageMySQL{
		db: db,
	}
}

// Create implements PostStorage
func (p *PostStorageMySQL) Create(ctx context.Context, post *model.Post) error {
	err := p.db.Create(post).Error
	return errors.Wrapf(err, "fail to create post data")
}

// FetchByPostID implements PostStorage
func (p *PostStorageMySQL) FetchByPostID(ctx context.Context, postID int64) (*model.Post, error) {
	postM := model.Post{}
	err := p.db.Model(&model.Post{}).Where("post_id = ?", postID).First(&postM).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "fail to query post %v", postID)
	}
	return &postM, nil
}

// HasPost implements PostStorage
func (p *PostStorageMySQL) HasPost(ctx context.Context, postID int64) (bool, error) {
	var count int64
	err := p.db.Model(&model.Post{}).
		Where("post_id = ?", postID).Count(&count).Error
	if err != nil {
		return false, errors.Wrap(err, "fails to check postID existence")
	}
	return count > 0, nil
}

// List implements PostStorage
func (p *PostStorageMySQL) List(ctx context.Context, order string, cursor string, cnt int) (list []*model.Post, newCursor string, err error) {
	cnt = funcs.Clip(cnt, 1, conf.Global.App.MaxPageSize)
	lastID, lastTime, err := decomposePageCursor(cursor)
	if err != nil {
		return nil, "", errors.Wrapf(err, "wrong cursor: %v", cursor)
	}

	var orderField string
	switch order {
	case PostOrderCreateTimeDesc:
		orderField = "created_at"
	case PostOrderReplyTimeDesc:
		orderField = "reply_time"
	default:
		return nil, "", errors.Errorf("unsupported post list order: %v", order)
	}

	err = p.db.Model(&model.Post{}).
		Where(fmt.Sprintf("%v <= ?", orderField), lastTime).
		Where("post_id < ?", lastID).
		Order(fmt.Sprintf("%v DESC", orderField)).
		Order("post_id DESC").
		Limit(cnt).
		Find(&list).Error
	if err != nil {
		return nil, "", errors.Wrap(err, "fail to query post")
	}
	if len(list) == 0 {
		return nil, cursor, nil
	}

	n := len(list)
	switch order {
	case PostOrderCreateTimeDesc:
		newCursor = composePageCursor(list[n-1].PostID, list[n-1].CreatedAt)
	case PostOrderReplyTimeDesc:
		newCursor = composePageCursor(list[n-1].PostID, list[n-1].ReplyTime)
	default:
		return nil, "", errors.Errorf("unsupported post list order (2): %v", order)
	}
	return list, newCursor, nil
}

// IncrementReplyNum implements PostStorage
func (p *PostStorageMySQL) IncrementReplyNum(ctx context.Context, postID int64, incr int) error {
	now := time.Now()
	err := p.db.Model(&model.Post{}).Where("post_id = ?", postID).
		Updates(map[string]interface{}{
			"reply_time": now,
			"updated_at": now,
			"reply_num":  gorm.Expr("reply_num + ?", incr),
		}).Error
	return errors.Wrapf(err, "fails to increment reply num")
}

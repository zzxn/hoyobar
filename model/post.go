package model

import "time"

type Post struct {
	Model
	PostID    int64     `gorm:"uniqueIndex;index:idx_reply_time_post_id,priority:2;index:idx_created_at_post_id,priority:2"`
	CreatedAt time.Time `gorm:"index:idx_created_at_post_id,priority:1"`
	ReplyTime time.Time `gorm:"index:idx_reply_time_post_id,priority:1"`
	ReplyNum  int64
	AuthorID  int64  `gorm:"index"`
	Title     string `gorm:"size:50"`
	Content   string
}

func (Post) TableName() string {
	return "post"
}

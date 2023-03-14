package model

import "time"

type Post struct {
	Model
	PostID    int64 `gorm:"uniqueIndex"`
	AuthorID  int64 `gorm:"index"`
	Title     string
	Content   string
	ReplyTime time.Time `gorm:"index"`
}

func (Post) TableName() string {
	return "post"
}

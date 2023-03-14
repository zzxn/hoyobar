package model

import "time"

type PostStat struct {
    Model
    PostID int64 `gorm:"uniqueIndex"`
    ReplyTime time.Time `gorm:"index"`
    ReplyNum int64
}

func (PostStat) TableName() string {
    return "post_stat"
}

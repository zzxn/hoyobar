package model

type PostReply struct {
	Model
	ReplyID  int64 `gorm:"uniqueIndex"`
	AuthorID int64 `gorm:"index"`
	PostID   int64 `gorm:"index"`
	Content  string
}

func (PostReply) TableName() string {
	return "post_reply"
}

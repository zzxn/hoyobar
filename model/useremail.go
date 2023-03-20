package model

import (
	"hoyobar/conf"
	"hoyobar/util/crypt"
	"strconv"

	"gorm.io/gorm"
)

type UserEmail struct {
	Model
	Email  string `gorm:"uniqueIndex;size:320"`
	UserID int64  `gorm:"uniqueIndex"`
}

func (UserEmail) TableName() string {
	return "user_email"
}

func TableOfUserEmail(useremail *UserEmail, email string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		shardIdx := crypt.HashString(email, int64(conf.Global.Sharding.UserShardN))
		tableName := useremail.TableName() + strconv.FormatInt(shardIdx, 10)
		return db.Table(tableName)
	}
}

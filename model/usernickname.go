package model

import (
	"hoyobar/conf"
	"hoyobar/util/crypt"
	"strconv"

	"gorm.io/gorm"
)

type UserNickname struct {
	Model
	Nickname string `gorm:"uniqueIndex;size:50"`
	UserID   int64  `gorm:"uniqueIndex"`
}

func (UserNickname) TableName() string {
	return "user_nickname"
}

func TableOfUserNickname(userNickname *UserNickname, nickname string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		shardIdx := crypt.HashString(nickname, int64(conf.Global.Sharding.UserShardN))
		tableName := userNickname.TableName() + strconv.FormatInt(shardIdx, 10)
		return db.Table(tableName)
	}
}

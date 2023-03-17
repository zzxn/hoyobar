package model

import (
	"hoyobar/conf"
	"hoyobar/util/crypt"
	"strconv"

	"gorm.io/gorm"
)

type UserPhone struct {
	Model
	Phone  string `gorm:"uniqueIndex;size:30"`
	UserID int64  `gorm:"uniqueIndex"`
}

func (UserPhone) TableName() string {
	return "user_phone"
}

func TableOfUserPhone(userphone *UserPhone, phone string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		shardIdx := crypt.HashString(phone, int64(conf.Global.Sharding.UserPhoneShardN))
		tableName := userphone.TableName() + strconv.FormatInt(shardIdx, 10)
		return db.Table(tableName)
	}
}

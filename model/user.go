package model

import (
	"database/sql"
	"hoyobar/conf"
	"hoyobar/util/myhash"
	"strconv"

	"gorm.io/gorm"
)

type User struct {
	Model
	UserID   int64          `gorm:"uniqueIndex"`
	Email    sql.NullString `gorm:"size:320"`
	Phone    sql.NullString `gorm:"size:30"`
	Nickname string         `gorm:"size:50"`
	Password string         `gorm:"size:100"`
}

func (User) TableName() string {
	return "user"
}

func TableOfUser(userID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(TableNameOfUser(userID))
	}
}

func TableNameOfUser(userID int64) string {
	shardIdx := myhash.HashSnowflakeID(userID, int64(conf.Global.Sharding.UserShardN))
	tableName := User{}.TableName() + strconv.FormatInt(shardIdx, 10)
	return tableName
}

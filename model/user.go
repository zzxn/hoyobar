package model

import (
	"database/sql"
	"fmt"
	"hoyobar/conf"
	"hoyobar/util/myhash"
	"hoyobar/util/regexes"
	"strconv"

	"gorm.io/gorm"
)

type User struct {
	Model
	UserID   int64          `gorm:"uniqueIndex"`
	Email    sql.NullString `gorm:"size:320"`
	Phone    sql.NullString `gorm:"size:30"`
	Nickname string
	Password string
}

func (User) TableName() string {
	return "user"
}

func TableOfUser(user *User, userID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		shardIdx := myhash.HashSnowflakeID(userID, int64(conf.Global.Sharding.UserShardN))
		tableName := user.TableName() + strconv.FormatInt(shardIdx, 10)
		return db.Table(tableName)
	}
}

func UsernameField(username string) (string, error) {
	if regexes.Email.MatchString(username) {
		return "email", nil
	}
	if regexes.Phone.MatchString(username) {
		return "phone", nil
	}
	return "", fmt.Errorf("%v is not a valid username", username)
}

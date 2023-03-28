package model

import (
	"hoyobar/conf"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Base type for model.
// time.Time should be parsed into:
// - mysql: DATETIME(3), see: https://github.com/go-gorm/mysql/blob/master/mysql.go#L401
// - sqlite: DATETIME. As sqlite has not seperate datetime type, it will be stored as string or interger.
// The pricision is determined by the gorm. (not sure about it, anyway it's not important, we just use it to debug)
type Model struct {
	ID        uint64         `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"index"`
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func Migrate(db *gorm.DB) {
	// TODO: do we need to do this?
	err := db.AutoMigrate(
		&Post{},
		&PostReply{},
	)
	if err != nil {
		panic(err)
	}

	// user need sharding
	// unique index name cannot be the same, why?
	autoMigrateShard(db, conf.Global.Sharding.UserShardN, User{})
	autoMigrateShard(db, conf.Global.Sharding.UserShardN, UserEmail{})
	autoMigrateShard(db, conf.Global.Sharding.UserShardN, UserPhone{})
	autoMigrateShard(db, conf.Global.Sharding.UserShardN, UserNickname{})
}

func autoMigrateShard(db *gorm.DB, shardN int, model interface{ TableName() string }) {
	tableName := model.TableName()
	for i := 0; i < shardN; i++ {
		err := db.Table(tableName + strconv.Itoa(i)).AutoMigrate(&model)
		if err != nil {
			panic(err)
		}
	}
}

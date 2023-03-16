package model

import (
	"hoyobar/conf"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

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

func Init(db *gorm.DB) {
	if conf.Global == nil {
		panic("conf.Global is nil")
	}
	if db == nil {
		panic("db is nil")
	}
	DB = db
	if conf.Global.DB.AutoMigrate {
		migrate(DB)
	}
}

func migrate(db *gorm.DB) {
	// TODO: do we need to do this?
	var err error
	err = db.Debug().AutoMigrate(
		&Post{},
		&PostStat{},
		&PostReply{},
	)
	if err != nil {
		panic(err)
	}

	// user need sharding
	// unique index name cannot be the same, why?
	autoMigrateShard(db, conf.Global.Sharding.UserShardN, User{}.TableName(), &User{})
	autoMigrateShard(db, conf.Global.Sharding.UserEmailShardN, UserEmail{}.TableName(), &UserEmail{})
	autoMigrateShard(db, conf.Global.Sharding.UserPhoneShardN, UserPhone{}.TableName(), &UserPhone{})
}

func autoMigrateShard(db *gorm.DB, shardN int, tableName string, modelPtr interface{}) {
	for i := 0; i < shardN; i++ {
		err := db.Debug().Table(tableName + strconv.Itoa(i)).AutoMigrate(modelPtr)
		if err != nil {
			panic(err)
		}
	}
}

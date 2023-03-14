package model

import (
	"time"

	"gorm.io/gorm"
)

var DB *gorm.DB

type Model struct {
	ID        uint64         `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"index"`
	UpdatedAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func Init(db *gorm.DB) {
	if db == nil {
		panic("db is nil")
	}
	DB = db
}

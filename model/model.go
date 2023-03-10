package model

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint64 `gorm:"primarykey"`
    CreatedAt time.Time `gorm:"index"`
    UpdatedAt time.Time `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}


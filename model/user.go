package model

import (
	"database/sql"
)

type User struct {
    Model
    UserID int64 `gorm:"uniqueIndex"`
    Email sql.NullString `gorm:"uniqueIndex"`
    Phone sql.NullString `gorm:"uniqueIndex"`
    Nickname string
    Password string
}

func (User) TableName() string {
    return "user"
}

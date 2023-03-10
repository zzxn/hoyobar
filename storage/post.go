package storage

import "gorm.io/gorm"

type PostStorage struct {
    DB *gorm.DB
}

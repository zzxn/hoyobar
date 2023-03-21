package storage

import "hoyobar/model"

type UserStorage interface {
	FetchUser(userID int64) (*model.User, error)
	HasUser(userID int64) (bool, error)
	PhoneToUserID(phone string) (int64, error)
	EmailToUserID(email string) (int64, error)
	NicknameToUserID(nickname string) (int64, error)
	CreateUser(user *model.User) error
}

type PostStorage interface {
}

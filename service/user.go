package service

import (
	"hoyobar/storage"
)


type UserService struct {
    UserStorage *storage.UserStorage
}

type UserBasic struct {
    UserID string
    Username string
    Password string
}

func (u *UserService) Verify(account string) {

}

func (u *UserService) Register(username string, password string, vcode string) *UserBasic {
    return nil
}

func (U *UserService) FetchInfoByUsername(username string) *UserBasic {
    return nil
}

func (U *UserService) FetchInfoByUserID(userID string) *UserBasic {
    return nil
}

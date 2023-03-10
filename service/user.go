package service

import (
	"database/sql"
	"fmt"
	"hoyobar/model"
)


type UserService struct {}

type UserBasic struct {
    UserID int64
    Phone string
    Email string
    Password string
}

func (u *UserService) Verify(account string) (bool, error) {
    return false, nil
}

func (u *UserService) Register(username string, password string, vcode string) (*UserBasic, error) {
    var err error
    var userID int64 = 11111
    userModel := model.User{
        UserID: userID,
        Phone: sql.NullString{String: username, Valid: true},
        Password: password,
        Nickname: "zzxn",
    }
    err = model.DB.Create(&userModel).Error
    if err != nil {
        return nil, fmt.Errorf("fail to create user: %v", err)
    }

    userModel = model.User{}
    err = model.DB.Where("user_id = ?", userID).First(&userModel).Error
    if err != nil {
        return nil, fmt.Errorf("fail to find user after creation: %v", err)
    }
    return &UserBasic{
        UserID: userModel.UserID, 
        Phone: userModel.Phone.String, 
        Password: userModel.Password,
    }, nil 
}

func (U *UserService) FetchInfoByUsername(username string) (*UserBasic, error) {
    return nil, nil
}

func (U *UserService) FetchInfoByUserID(userID string) (*UserBasic, error) {
    return nil, nil
}

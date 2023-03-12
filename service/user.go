package service

import (
	"database/sql"
	"fmt"
	"hoyobar/model"
	"hoyobar/util/idgen"
	"strconv"
	"sync"
)


type UserService struct {
    authToken2UserID struct {
        data map[string]int64
        mu sync.RWMutex
    }
}

type UserBasic struct {
    UserID int64
    Phone string
    Email string
    Nickname string
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
        Nickname: userModel.Nickname,
    }, nil 
}

func (u *UserService) GenAuthToken(userID int64) string {
    // TODO: store auth token into redis
    token := strconv.FormatInt(idgen.New(), 10)
    u.authToken2UserID.mu.Lock()
    defer u.authToken2UserID.mu.Unlock()
    u.authToken2UserID.data[token] = userID
    return token 
}

func (u *UserService) AuthTokenToUserID(authToken string) (userID int64, err error) {
    u.authToken2UserID.mu.RLock()
    defer u.authToken2UserID.mu.RUnlock()
    userID, ok := u.authToken2UserID.data[authToken]
    if !ok {
        return 0, fmt.Errorf("auth token not found")
    }
    return userID, nil
}

func (u *UserService) FetchInfoByUsername(username string) (*UserBasic, error) {
    return nil, nil
}

func (u *UserService) FetchInfoByUserID(userID string) (*UserBasic, error) {
    return nil, nil
}

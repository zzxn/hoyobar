package service

import (
	"database/sql"
	"fmt"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/util/crypt"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"log"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	cache mycache.Cache
}

func NewUserService(cache mycache.Cache) *UserService {
	if conf.Global == nil {
		log.Fatalf("conf.Global is not initialized")
	}
	userService := &UserService{
		cache: cache,
	}
	return userService
}

type UserBasic struct {
	UserID    int64 `json:"user_id,string"`
	Phone     string
	Email     string
	Nickname  string
	AuthToken string
}

type RegisterInfo struct {
    Username string
    Password string
    Vcode string
    Nickname string
}

// send verification code to email/phone represented by username
func (u *UserService) Verify(username string) error {
    // TODO: 
	return nil
}

// check if username and verification code is matched
func (u *UserService) checkVcode(username string, vcode string) (bool, error) {
    return true, nil
}

func (u *UserService) Register(args *RegisterInfo) (*UserBasic, error) {
	var err error
    username, rawPass := args.Username, args.Password

    if !crypt.CheckPasswordStrength(rawPass) {
        return nil, myerr.ErrWeakPassword
    }

    vcodeOK, err := u.checkVcode(username, args.Vcode)
    if err != nil {
        return nil, err
    }
    if !vcodeOK {
        return nil, myerr.ErrWrongVcode
    }

	// TODO fixme: here we assume username is phone
	userExist, err := u.ExistByUsername(username)
	if err != nil {
		return nil, err
	}
	if userExist {
		return nil, myerr.ErrDupUser
    }

	var userID int64 = idgen.New()
	passhash, err := crypt.HashPassword(rawPass)
	if err != nil {
		return nil, myerr.NewOtherErr(err, "fail to hash password")
	}
	userModel := model.User{
		UserID:   userID,
		Phone:    sql.NullString{String: username, Valid: true},
		Password: passhash,
		Nickname: args.Nickname,
	}
	err = model.DB.Create(&userModel).Error
	if err != nil {
		return nil, myerr.NewOtherErr(err, "fail to create user %q", username)
	}

	userModel = model.User{}
	err = model.DB.Where("user_id = ?", userID).First(&userModel).Error
	if err != nil {
		return nil, myerr.NewOtherErr(err, "fail to find user %q after creation", username)
	}

	authToken := u.GenAndStoreAuthToken(userID)
	return &UserBasic{
		UserID:    userModel.UserID,
		Phone:     userModel.Phone.String,
		Nickname:  userModel.Nickname,
		AuthToken: authToken,
	}, nil
}

func authTokenToCacheKey(authToken string) string {
	return "(auth_token)" + authToken
}

func (u *UserService) GenAndStoreAuthToken(userID int64) string {
	// TODO: store auth token into redis
	token := strings.ReplaceAll(uuid.NewString(), "-", "")
	u.cache.Set(authTokenToCacheKey(token), userID, conf.Global.App.AuthTokenExpire)
	return token
}

// convert auth token to user ID, also refresh cache
func (u *UserService) AuthTokenToUserID(authToken string) (userID int64, err error) {
	key := authTokenToCacheKey(authToken)

	// get user ID from cache
	value, err := u.cache.Get(key)
	if err != nil {
		return 0, myerr.NewOtherErr(err, "fail to query auth token cache key %q", key)
	}
	if value == nil {
		return 0, myerr.ErrNotLogin
	}
	userID, ok := value.(int64)
	if !ok {
		return 0, myerr.ErrOther.Wrap(fmt.Errorf("type of auth token cache value expect int64, got %T", value))
	}

	// update cache to avoid expiration, ignore error
	_ = u.cache.Set(key, userID, conf.Global.App.AuthTokenExpire)
	return userID, nil
}

func (u *UserService) Login(username, password string) (*UserBasic, error) {
	var err error
	userModel := model.User{}
	err = model.DB.Where("phone = ?", username).First(&userModel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, myerr.ErrUserNotFound
	}
	if err != nil {
		return nil, myerr.NewOtherErr(err, "fail to find user")
	}
	if false == crypt.CompareHashAndPassword(userModel.Password, password) {
		return nil, myerr.ErrWrongPassword
	}
	authToken := u.GenAndStoreAuthToken(userModel.UserID)
	return &UserBasic{
		UserID:    userModel.UserID,
		Phone:     userModel.Phone.String,
		Nickname:  userModel.Nickname,
		AuthToken: authToken,
	}, nil
}

// check if username (email/phone) exist in system.
func (u *UserService) ExistByUsername(username string) (bool, error) {
	var count int64
	// TODO: fix here, check it's phone or email
	err := model.DB.Model(&model.User{}).Where("phone = ?", username).Count(&count).Error
	if err != nil {
		return true, myerr.NewOtherErr(err, "fail to check user existence")
	}
	return count > 0, nil
}

// func (u *UserService) FetchInfoByUsername(username string) (*UserBasic, error) {
//     return nil, nil
// }

// func (u *UserService) FetchInfoByUserID(userID string) (*UserBasic, error) {
//     return nil, nil
// }

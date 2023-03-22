package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/storage"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/mycache/keys"
	"hoyobar/util/myerr"
	"hoyobar/util/myhash"
	"hoyobar/util/regexes"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UserService struct {
	cache       mycache.Cache
	userStorage storage.UserStorage
}

func NewUserService(
	cache mycache.Cache,
	userStorage storage.UserStorage,
) *UserService {
	userService := &UserService{
		cache:       cache,
		userStorage: userStorage,
	}
	return userService
}

type UserBasic struct {
	UserID    int64  `json:"user_id,string"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	AuthToken string `json:"auth_token"`
}

type RegisterInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Vcode    string `json:"vcode"`
	Nickname string `json:"nickname"`
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

	if !regexes.Password.MatchString(rawPass) {
		return nil, myerr.ErrWeakPassword
	}
	passhash, err := myhash.HashPassword(rawPass)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to hash password")
	}

	vcodeOK, err := u.checkVcode(username, args.Vcode)
	if err != nil {
		return nil, err
	}
	if !vcodeOK {
		return nil, myerr.ErrWrongVcode
	}

	existUserID, err := u.NicknameToUserID(args.Nickname)
	if err != nil {
		return nil, err
	}
	if existUserID != 0 {
		return nil, myerr.ErrDupUser.WithEmsg("昵称已被占用")
	}

	existUserID, err = u.UsernameToUserID(username)
	if err != nil {
		return nil, err
	}
	if existUserID != 0 {
		return nil, myerr.ErrDupUser.WithEmsg("该账户已存在")
	}

	var userID int64 = idgen.New()

	usernameType := UsernameType(username)
	userModel := model.User{
		UserID:   userID,
		Password: passhash,
		Nickname: args.Nickname,
	}
	if usernameType == "phone" {
		userModel.Phone = sql.NullString{String: username, Valid: true}
	} else if usernameType == "email" {
		userModel.Email = sql.NullString{String: username, Valid: true}
	} else {
		return nil, myerr.ErrOther.WithEmsg("账号不是合法的邮箱或11位手机号")
	}

	err = u.userStorage.Create(&userModel)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to create user %q", username).
			WithEmsg("注册失败")
	}

	userInfoExpire := conf.Global.App.Expire.UserInfo

	if userModel.Email.Valid {
		_ = u.cache.SetInt64(context.TODO(),
			keys.EmailToUserID(userModel.Email.String), userID,
			userInfoExpire,
		)
	}

	if userModel.Phone.Valid {
		_ = u.cache.SetInt64(context.TODO(),
			keys.PhoneToUserID(userModel.Phone.String), userID,
			userInfoExpire,
		)
	}

	_ = u.cache.SetInt64(context.TODO(),
		keys.NicknameToUserID(userModel.Nickname), userID,
		userInfoExpire,
	)

	authToken, err := u.genAndStoreAuthToken(userID)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to write auth token").WithEmsg("请稍后尝试登录")
	}

	userBasic := &UserBasic{
		UserID:    userModel.UserID,
		Phone:     userModel.Phone.String,
		Email:     userModel.Email.String,
		Nickname:  userModel.Nickname,
		AuthToken: authToken,
	}

	u.writeCacheUserBasic(*userBasic)
	return userBasic, nil
}

func (u *UserService) genAndStoreAuthToken(userID int64) (string, error) {
	token := strings.ReplaceAll(uuid.NewString(), "-", "")
	key := keys.AuthToken(token)
	expire := conf.Global.App.Expire.AuthToken
	if err := u.cache.SetInt64(context.TODO(), key, userID, expire); err != nil {
		return "", errors.Wrapf(err, "fail to write auth token to redis")
	}
	return token, nil
}

func (u *UserService) writeCacheUserBasic(user UserBasic) {
	if user.UserID == 0 {
		return
	}
	key := keys.UserBasic(user.UserID)
	user.AuthToken = ""
	value, err := json.Marshal(user)
	if err != nil {
		return
	}
	expire := conf.Global.App.Expire.UserInfo
	_ = u.cache.Set(context.TODO(), key, string(value), expire)
}

func (u *UserService) readCacheUserBasic(userID int64) *UserBasic {
	key := keys.UserBasic(userID)
	data, err := u.cache.Get(context.TODO(), key)
	if err != nil {
		return nil
	}
	value := &UserBasic{}
	err = json.Unmarshal([]byte(data), value)
	if err != nil {
		log.Printf("fail to parse cache value with key %v", key)
		return nil
	}
	return value
}

// convert auth token to user ID, also refresh cache
func (u *UserService) AuthTokenToUserID(authToken string) (userID int64, err error) {
	key := keys.AuthToken(authToken)

	// get user ID from cache
	userID, err = u.cache.GetInt64(context.TODO(), key)
	if err == mycache.ErrNotFound {
		return 0, myerr.ErrNotLogin
	}
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to query auth token cache key %q", key)
	}
	return userID, nil
}

func (u *UserService) Login(username, password string) (*UserBasic, error) {
	var err error

	userID, err := u.UsernameToUserID(username)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fails to query username %v", username)
	}
	if userID == 0 {
		return nil, myerr.ErrUserNotFound
	}

	var userBasic *UserBasic
	userBasic = u.readCacheUserBasic(userID)
	if userBasic == nil {
		userModel, err := u.userStorage.FetchByUserID(userID)
		if err != nil {
			return nil, myerr.OtherErrWarpf(err, "fail to find user")
		}
		if userModel == nil {
			return nil, myerr.ErrOther.WithEmsg("未找到用户数据，请联系客服")
		}

		if false == myhash.CompareHashAndPassword(userModel.Password, password) {
			return nil, myerr.ErrWrongPassword
		}
		userBasic = &UserBasic{
			UserID:   userModel.UserID,
			Phone:    userModel.Phone.String,
			Email:    userModel.Email.String,
			Nickname: userModel.Nickname,
		}
	}
	authToken, err := u.genAndStoreAuthToken(userBasic.UserID)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to write auth token").WithEmsg("请稍后尝试登录")
	}
	userBasic.AuthToken = authToken
	return userBasic, nil
}

// transform username to user ID.
// attenion: if username not exist, return 0, nil
func (u *UserService) UsernameToUserID(username string) (int64, error) {
	var userID int64
	var err error

	usernameType := UsernameType(username)
	expire := conf.Global.App.Expire.UserInfo

	if usernameType == "phone" {
		key := keys.PhoneToUserID(username)
		userID, err = u.cache.GetInt64(context.TODO(), key)
		if err == nil {
			log.Printf("success query phone %v from cache\n", username)
			return userID, err
		}
		err = nil
		userID, err = u.userStorage.PhoneToUserID(username)
		if err == nil && userID != 0 {
			_ = u.cache.SetInt64(context.TODO(), key, userID, expire)
		}
	} else if usernameType == "email" {
		key := keys.EmailToUserID(username)
		userID, err = u.cache.GetInt64(context.TODO(), key)
		if err == nil {
			log.Printf("success query email %v from cache\n", username)
			return userID, err
		}
		err = nil
		userID, err = u.userStorage.EmailToUserID(username)
		if err == nil && userID != 0 {
			_ = u.cache.SetInt64(context.TODO(), key, userID, expire)
		}
	} else {
		return 0, myerr.OtherErrWarpf(fmt.Errorf(""), "not support username type %v", usernameType).
			WithEmsg("账号不是合法的邮箱或手机号")
	}

	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to check username existence")
	}
	return userID, nil
}

// transform nickname to user ID.
// attenion: if nickname not exist, return 0, nil
func (u *UserService) NicknameToUserID(nickname string) (int64, error) {
	key := keys.NicknameToUserID(nickname)
	userID, err := u.cache.GetInt64(context.TODO(), key)
	if err == nil {
		log.Printf("success query nickname %v from cache\n", nickname)
		return userID, err
	}
	err = nil
	userID, err = u.userStorage.NicknameToUserID(nickname)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to check nickname existence")
	}
	if userID != 0 {
		expire := conf.Global.App.Expire.UserInfo
		_ = u.cache.SetInt64(context.TODO(), key, userID, expire)
	}
	return userID, nil
}

func UsernameType(username string) string {
	if regexes.Email.MatchString(username) {
		return "email"
	}
	if regexes.Phone.MatchString(username) {
		return "phone"
	}
	return "unknown"
}

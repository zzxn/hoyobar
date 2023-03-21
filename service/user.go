package service

import (
	"database/sql"
	"fmt"
	"hoyobar/conf"
	"hoyobar/model"
	"hoyobar/storage"
	"hoyobar/util/idgen"
	"hoyobar/util/mycache"
	"hoyobar/util/myerr"
	"hoyobar/util/myhash"
	"hoyobar/util/regexes"
	"strings"

	"github.com/google/uuid"
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

	authToken := u.GenAndStoreAuthToken(userID)
	return &UserBasic{
		UserID:    userModel.UserID,
		Phone:     userModel.Phone.String,
		Email:     userModel.Email.String,
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
		return 0, myerr.OtherErrWarpf(err, "fail to query auth token cache key %q", key)
	}
	if value == nil {
		return 0, myerr.ErrNotLogin
	}
	userID, ok := value.(int64)
	if !ok {
		return 0, myerr.ErrOther.WithCause(fmt.Errorf("type of auth token cache value expect int64, got %T", value))
	}

	// update cache to avoid expiration, ignore error
	_ = u.cache.Set(key, userID, conf.Global.App.AuthTokenExpire)
	return userID, nil
}

func (u *UserService) Login(username, password string) (*UserBasic, error) {
	var err error
	var userModel *model.User = nil

	userID, err := u.UsernameToUserID(username)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fails to query username %v", username)
	}
	if userID == 0 {
		return nil, myerr.ErrUserNotFound
	}

	userModel, err = u.userStorage.FetchByUserID(userID)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to find user")
	}
	if userModel == nil {
		return nil, myerr.ErrOther.WithEmsg("未找到用户数据，请联系客服")
	}

	if false == myhash.CompareHashAndPassword(userModel.Password, password) {
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

// transform username to user ID.
// attenion: if username not exist, return 0, nil
func (u *UserService) UsernameToUserID(username string) (int64, error) {
	var userID int64
	var err error

	usernameType := UsernameType(username)

	if usernameType == "phone" {
		userID, err = u.userStorage.PhoneToUserID(username)
	} else if usernameType == "email" {
		userID, err = u.userStorage.EmailToUserID(username)
	} else {
		return 0, myerr.OtherErrWarpf(fmt.Errorf(""), "not support username type %v", usernameType).
			WithEmsg("账号不是合法的邮箱或手机号")
	}

	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to check user existence")
	}
	return userID, nil
}

// transform nickname to user ID.
// attenion: if nickname not exist, return 0, nil
func (u *UserService) NicknameToUserID(nickname string) (int64, error) {
	userID, err := u.userStorage.NicknameToUserID(nickname)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to check user existence")
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

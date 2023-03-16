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
	"hoyobar/util/regexes"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	cache mycache.Cache
}

func NewUserService(cache mycache.Cache) *UserService {
	userService := &UserService{
		cache: cache,
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
	passhash, err := crypt.HashPassword(rawPass)
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

	existUserID, err := u.UsernameToUserID(username)
	if err != nil {
		return nil, err
	}
	if existUserID != 0 {
		return nil, myerr.ErrDupUser
	}

	var userID int64 = idgen.New()
	err = u.createUsername(username, userID)
	if err != nil {
		return nil, err
	}

	usernameField, err := model.UsernameField(username)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "%v not a valid username", username).
			WithEmsg("账号不是合法的邮箱或11位手机号")
	}
	userModel := model.User{
		UserID:   userID,
		Password: passhash,
		Nickname: args.Nickname,
	}
	if usernameField == "phone" {
		userModel.Phone = sql.NullString{String: username, Valid: true}
	} else if usernameField == "email" {
		userModel.Email = sql.NullString{String: username, Valid: true}
	} else {
		return nil, myerr.ErrOther.WithEmsg("不支持的账号类型")
	}

	err = model.DB.Scopes(model.TableOfUser(&userModel, userID)).Create(&userModel).Error
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to create user %q", username).
			WithEmsg("注册异常，请联系管理员")
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
	userModel := model.User{}

	// TODO: query user id
	userID, err := u.UsernameToUserID(username)
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fails to query username %v", username)
	}
	if userID == 0 {
		return nil, myerr.ErrUserNotFound
	}

	err = model.DB.Scopes(model.TableOfUser(&userModel, userID)).
		Where("user_id = ?", userID).First(&userModel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, myerr.ErrUserNotFound.WithEmsg("用户数据异常，请联系客服")
	}
	if err != nil {
		return nil, myerr.OtherErrWarpf(err, "fail to find user")
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

// transform username to user ID.
// attenion: if username not exist, return 0, nil
func (u *UserService) UsernameToUserID(username string) (int64, error) {
	var userID int64
	var err error

	usernameField, err := model.UsernameField(username)
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "%v not a valid username", username).
			WithEmsg("账号不是合法的邮箱或手机号")
	}

	if usernameField == "phone" {
		userPhoneM := model.UserPhone{}
		err = model.DB.Scopes(model.TableOfUserPhone(&userPhoneM, username)).
			Where("phone = ?", username).First(&userPhoneM).Error
		userID = userPhoneM.UserID
	} else if usernameField == "email" {
		userEmailM := model.UserEmail{}
		err = model.DB.Scopes(model.TableOfUserEmail(&userEmailM, username)).
			Where("email = ?", username).First(&userEmailM).Error
		userID = userEmailM.UserID
	} else {
		return 0, myerr.OtherErrWarpf(fmt.Errorf(""), "not support username typed %v", usernameField).
			WithEmsg("不支持的账户类型")
	}

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, myerr.OtherErrWarpf(err, "fail to check user existence")
	}
	return userID, nil
}

func (u *UserService) createUsername(username string, userID int64) error {
	var err error

	usernameField, err := model.UsernameField(username)
	if err != nil {
		return myerr.OtherErrWarpf(err, "%v not a valid username", username)
	}

	if usernameField == "phone" {
		err = model.DB.Scopes(model.TableOfUserPhone(&model.UserPhone{}, username)).
			Create(&model.UserPhone{Phone: username, UserID: userID}).Error
	} else if usernameField == "email" {
		err = model.DB.Scopes(model.TableOfUserEmail(&model.UserEmail{}, username)).
			Create(&model.UserEmail{Email: username, UserID: userID}).Error
	} else {
		return myerr.OtherErrWarpf(fmt.Errorf(""), "not support username typed %v", usernameField).
			WithEmsg("不支持的账户类型")
	}

	if err != nil {
		return myerr.OtherErrWarpf(err, "fail to create username")
	}
	return nil
}

// func (u *UserService) FetchInfoByUsername(username string) (*UserBasic, error) {
//     return nil, nil
// }

// func (u *UserService) FetchInfoByUserID(userID string) (*UserBasic, error) {
//     return nil, nil
// }

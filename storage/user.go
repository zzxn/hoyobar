package storage

import (
	"context"
	"hoyobar/model"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UserStorageMySQL struct {
	db *gorm.DB
}

var _ = UserStorage(new(UserStorageMySQL))

func NewUserStorageMySQL(db *gorm.DB) *UserStorageMySQL {
	return &UserStorageMySQL{
		db: db,
	}
}

// FetchUser implements UserStorage
func (u *UserStorageMySQL) FetchByUserID(ctx context.Context, userID int64) (*model.User, error) {
	var userModel model.User
	err := u.db.Scopes(model.TableOfUser(&userModel, userID)).
		Where("user_id = ?", userID).First(&userModel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "fail to fetch user")
	}
	return &userModel, nil
}

// HasUser implements UserStorage
func (u *UserStorageMySQL) HasUser(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := u.db.Scopes(model.TableOfUser(&model.User{}, userID)).
		Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return false, errors.Wrap(err, "fails to check user existence")
	}
	return count > 0, nil
}

// CreateUser implements UserStorage
func (u *UserStorageMySQL) Create(ctx context.Context, user *model.User) error {
	var err error
	userID := user.UserID

	err = u.createNickname(user.Nickname, userID)
	if err != nil {
		return errors.Wrapf(err, "fail to create nickname for userID=%v", userID)
	}

	if user.Phone.Valid {
		err = u.createPhone(user.Phone.String, userID)
		if err != nil {
			return errors.Wrapf(err,
				"fail to create phone for userID=%v, but nickname success", userID,
			)
		}
	}

	if user.Email.Valid {
		err = u.createEmail(user.Email.String, userID)
		if err != nil {
			return errors.Wrapf(err,
				"fail to create email for userID=%v, but phone/nickname success", userID,
			)
		}
	}

	err = u.db.Scopes(model.TableOfUser(user, userID)).Create(user).Error
	if err != nil {
		return errors.Wrapf(err,
			"fail to create user for userID=%v, but nickname/phone/email success", userID,
		)
	}
	return nil
}

// PhoneToUserID implements UserStorage
func (u *UserStorageMySQL) PhoneToUserID(ctx context.Context, phone string) (int64, error) {
	var userID int64
	var err error

	userPhoneM := model.UserPhone{}
	err = u.db.Scopes(model.TableOfUserPhone(&userPhoneM, phone)).
		Where("phone = ?", phone).First(&userPhoneM).Error
	userID = userPhoneM.UserID

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Errorf("fails to query user phone: %v", err)
	}
	return userID, nil
}

// EmailToUserID implements UserStorage
func (u *UserStorageMySQL) EmailToUserID(ctx context.Context, email string) (int64, error) {
	var userID int64
	var err error

	userEmailM := model.UserEmail{}
	err = u.db.Scopes(model.TableOfUserEmail(&userEmailM, email)).
		Where("email = ?", email).First(&userEmailM).Error
	userID = userEmailM.UserID

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "fails to query user email")
	}
	return userID, nil
}

// NicknameToUserID implements UserStorage
func (u *UserStorageMySQL) NicknameToUserID(ctx context.Context, nickname string) (int64, error) {
	var userID int64
	var err error

	nicknameM := model.UserNickname{}
	err = u.db.Scopes(model.TableOfUserNickname(&nicknameM, nickname)).
		Where("nickname = ?", nickname).First(&nicknameM).Error
	userID = nicknameM.UserID

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrapf(err, "fails to query user nickname")
	}
	return userID, nil
}

func (u *UserStorageMySQL) createPhone(phone string, userID int64) error {
	var err error
	err = u.db.Scopes(model.TableOfUserPhone(&model.UserPhone{}, phone)).
		Create(&model.UserPhone{Phone: phone, UserID: userID}).Error
	return errors.Wrapf(err, "fails to create phone")
}

func (u *UserStorageMySQL) createEmail(email string, userID int64) error {
	var err error
	err = u.db.Scopes(model.TableOfUserEmail(&model.UserEmail{}, email)).
		Create(&model.UserEmail{Email: email, UserID: userID}).Error
	return errors.Wrapf(err, "fails to create email")
}

func (u *UserStorageMySQL) createNickname(nickname string, userID int64) error {
	var err error
	err = u.db.Scopes(model.TableOfUserNickname(&model.UserNickname{}, nickname)).
		Create(&model.UserNickname{Nickname: nickname, UserID: userID}).Error
	return errors.Wrapf(err, "fails to create nickname")
}

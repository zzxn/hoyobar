package storage

import (
	"context"
	"hoyobar/conf"
	"hoyobar/model"
	"log"
	"runtime/debug"

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
	err := u.db.Scopes(model.TableOfUser(userID)).
		Where("user_id = ?", userID).First(&userModel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "fail to fetch user")
	}
	return &userModel, nil
}

// BatchFetchByUserIDs implements UserStorage
func (u *UserStorageMySQL) BatchFetchByUserIDs(ctx context.Context, userIDs []int64, fields []string) ([]*model.User, error) {
	userIDInFields := false
	for _, v := range fields {
		if v == "user_id" {
			userIDInFields = true
			break
		}
	}
	if !userIDInFields {
		fields = append([]string{"user_id"}, fields...)
	}

	nWorker := conf.Global.App.BatchFetchNWorker

	tableToUserIDs := make(map[string][]int64)
	for _, id := range userIDs {
		table := model.TableNameOfUser(id)
		tableToUserIDs[table] = append(tableToUserIDs[table], id)
	}

	taskChan := make(chan []int64, len(tableToUserIDs))
	for _, userIDs := range tableToUserIDs {
		taskChan <- userIDs
	}
	errChan := make(chan error, nWorker)
	resChan := make(chan []*model.User, len(tableToUserIDs))
	cancelChan := make(chan interface{})
	for i := 0; i < nWorker; i++ {
		go func(workerID int) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("BatchFetchByUserIDs worker %v panic %v\n", workerID, err)
					log.Printf("BatchFetchByUserIDs worker %v panic stack %v\n", workerID, debug.Stack())
					errChan <- errors.Errorf("worker %v panic", workerID)
				}
			}()
			for {
				select {
				case <-cancelChan:
					log.Printf("BatchFetchByUserIDs worker %v canceled", workerID)
					break
				default: // do nothing
				}
				userIDs, ok := <-taskChan
				if !ok {
					break
				}
				// must: len(userIDs) > 0
				userMs, err := u.batchFetchByUserIDsInOneTable(ctx, userIDs, fields)
				if err != nil {
					log.Printf("BatchFetchByUserIDs worker %v got err: %v\n", workerID, err)
					errChan <- err
					break
				}
				resChan <- userMs
			}
			log.Printf("BatchFetchByUserIDs worker %v exit", workerID)
		}(i)
	}

	idToIndices := make(map[int64][]int)
	for i, v := range userIDs {
		idToIndices[v] = append(idToIndices[v], i)
	}

	userMs := make([]*model.User, len(userIDs))
	for range tableToUserIDs {
		select {
		case err := <-errChan:
			close(cancelChan) // broadcast cancel signal
			err = errors.Wrapf(err, "fails to batch fetch users")
			return nil, err
		case userSubSet := <-resChan:
			for _, user := range userSubSet {
				for _, idx := range idToIndices[user.UserID] {
					userMs[idx] = copyUserModel(user)
				}
			}
		case <-ctx.Done():
			return nil, errors.Wrapf(ctx.Err(), "context canceled")
		}
	}

	return userMs, nil
}

func copyUserModel(user *model.User) *model.User {
	return &model.User{
		Model: model.Model{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
		},
		UserID:   user.UserID,
		Email:    user.Email,
		Phone:    user.Phone,
		Nickname: user.Nickname,
		Password: user.Password,
	}

}

// returns a set of users (no duplication).
// if a userID is not found, the user will not be present.
func (u *UserStorageMySQL) batchFetchByUserIDsInOneTable(
	ctx context.Context, userIDs []int64, fields []string,
) ([]*model.User, error) {
	userSet := make([]*model.User, 0, len(userIDs))
	err := u.db.Scopes(model.TableOfUser(userIDs[0])).
		Select(fields).
		Where("user_id in ?", userIDs).Find(&userSet).Error
	if err != nil {
		return nil, err
	}
	return userSet, nil

}

// HasUser implements UserStorage
func (u *UserStorageMySQL) HasUser(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := u.db.Scopes(model.TableOfUser(userID)).
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

	err = u.db.Scopes(model.TableOfUser(userID)).Create(user).Error
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
	err := u.db.Scopes(model.TableOfUserPhone(&model.UserPhone{}, phone)).
		Create(&model.UserPhone{Phone: phone, UserID: userID}).Error
	return errors.Wrapf(err, "fails to create phone")
}

func (u *UserStorageMySQL) createEmail(email string, userID int64) error {
	err := u.db.Scopes(model.TableOfUserEmail(&model.UserEmail{}, email)).
		Create(&model.UserEmail{Email: email, UserID: userID}).Error
	return errors.Wrapf(err, "fails to create email")
}

func (u *UserStorageMySQL) createNickname(nickname string, userID int64) error {
	err := u.db.Scopes(model.TableOfUserNickname(&model.UserNickname{}, nickname)).
		Create(&model.UserNickname{Nickname: nickname, UserID: userID}).Error
	return errors.Wrapf(err, "fails to create nickname")
}

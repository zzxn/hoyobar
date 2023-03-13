package myerr

import "github.com/pkg/errors"

type MyError struct {
	Ecode string
	Emsg  string
	cause error
}

// -1: 未知错误
// 2xxx：请求/参数格式错误
// 3xxx：权限/认证错误
// 4xxx: 其他错误

var (
	ErrUnknown       = newError("-1", "未知错误") // unrecognized error by biz logic
	ErrFailBindJSON  = newError("2001", "请求格式错误")
	ErrWrongPassword = newError("3001", "用户名或密码错误")
	ErrNotLogin      = newError("3002", "未登录")
	ErrOther         = newError("4000", "服务器内部错误") // 通用的其他错误
	ErrDuplicateUser = newError("4002", "该用户已存在")
	ErrUserNotFound  = newError("4003", "该用户不存在")
)

func (e *MyError) Error() string {
	return e.Emsg
}

func (e *MyError) Cause() error {
	return e.cause
}

func NewOtherErr(cause error, msg string, args ...interface{}) *MyError {
	return ErrOther.Wrap(errors.Wrapf(cause, msg, args))
}

func (e *MyError) Wrap(cause error) *MyError {
	return &MyError{
		Ecode: e.Ecode,
		Emsg:  e.Emsg,
		cause: cause,
	}
}

func newError(ecode string, emsg string) *MyError {
	return &MyError{
		Ecode: ecode,
		Emsg:  emsg,
	}
}

package myerr

import "github.com/pkg/errors"

type MyError struct {
	Ecode string
	Emsg  string
	cause error
}

// -1: 未知错误
// 1xxx：请求/参数格式错误
// 2xxx：权限/认证错误
// 3xxx: 其他错误

var (
	ErrUnknown = newError("-1", "未知错误") // unrecognized error by biz logic

	ErrBadReqBody   = newError("1000", "请求格式错误")
	ErrWeakPassword = newError("1001", "密码需要包含[数字]/[英文]/[其他字符]中的两种及以上，长度6-20")

	ErrAuth          = newError("2000", "权限/认证错误")
	ErrWrongPassword = newError("2001", "用户名或密码错误")
	ErrNotLogin      = newError("2002", "未登录")
	ErrWrongVcode    = newError("2003", "验证码错误")

	ErrOther            = newError("3000", "服务器内部错误") // 通用的其他错误
	ErrDupUser          = newError("3001", "该用户已存在")
	ErrUserNotFound     = newError("3002", "该用户不存在")
	ErrResourceNotFound = newError("3003", "该资源不存在")
	ErrNoMoreEntry      = newError("3004", "没有更多数据了")
	ErrTimeout          = newError("3005", "请求超时")
)

func (e *MyError) Error() string {
	return e.Emsg
}

func (e *MyError) Cause() error {
	return e.cause
}

// wrap a cause error with ErrOther.
// imsg is for inner message print.
func OtherErrWarpf(cause error, imsg string, args ...interface{}) *MyError {
	return ErrOther.WithCause(errors.Wrapf(cause, imsg, args))
}

func (e *MyError) WithCause(cause error) *MyError {
	return &MyError{
		Ecode: e.Ecode,
		Emsg:  e.Emsg,
		cause: cause,
	}
}

// return a new same-type of *MyError with new emsg.
// emsg is for user.
func (e *MyError) WithEmsg(emsg string) *MyError {
	return &MyError{
		Ecode: e.Ecode,
		Emsg:  emsg,
		cause: e.cause,
	}
}

func newError(ecode string, emsg string) *MyError {
	return &MyError{
		Ecode: ecode,
		Emsg:  emsg,
	}
}

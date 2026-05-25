package srverr

import (
	"errors"
	"fmt"
)

type ErrCode int

const (
	ErrCodeInternal ErrCode = iota + 1
	ErrCodeValidation
	ErrCodeNotFound
	ErrCodeAuthenticationFailure
	ErrCodeAuthorizationFailure
	ErrCodeTimeout
	ErrCodeStateConflict
	ErrCodePurged
	ErrCodeRateLimited
	ErrCodePreconditionFailed
	ErrCodeDependencyFailure
)

type Error struct {
	code    ErrCode
	message string
	err     error
}

func New(code ErrCode, message string, err error) *Error {
	return &Error{
		code:    code,
		message: message,
		err:     err,
	}
}

func Newf(code ErrCode, message string, errmsg string, args ...any) *Error {
	return New(code, message, fmt.Errorf(errmsg, args...))
}

func Unwrap(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var e *Error
	if !errors.As(err, &e) {
		return nil, false
	}
	return e, true
}

func ErrorIs(err error, code ErrCode) bool {
	e, ok := Unwrap(err)
	if !ok {
		return false
	}

	if e.Code() == code {
		return true
	}

	return false
}

func (e *Error) Code() ErrCode {
	return e.code
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Err() error {
	return e.err
}

func (e *Error) Error() string {
	str := fmt.Sprintf("code: %d, message: %s", e.code, e.message)
	if e.err != nil {
		str += fmt.Sprintf(", err: %v", e.err)
	}
	return str
}

func (e *Error) Unwrap() error {
	return e.err
}

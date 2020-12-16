package errors

import "github.com/pkg/errors"

const ErrCodeUnknown = -1

const (
	ErrCodeSaveFailed = iota
	ErrCodeNotFound
	ErrCodeQueryFailed
)

func Code(err error) int {
	if se := new(StatusError); errors.As(err, &se) {
		return se.Code
	}
	return ErrCodeUnknown
}

func SaveFailed(message string) *StatusError {
	return &StatusError{
		Code:    ErrCodeSaveFailed,
		Message: message,
	}
}

func isCode(err error, code int) bool {
	if se := new(StatusError); errors.As(err, &se) {
		return se.Code == code
	}
	return false
}

func IsSaveFailed(err error) bool {
	return isCode(err, ErrCodeSaveFailed)
}

func NotFound(message string) *StatusError {
	return &StatusError{
		Code:    ErrCodeNotFound,
		Message: message,
	}
}

func IsNotFound(err error) bool {
	return isCode(err, ErrCodeNotFound)
}

func QueryFailed(message string) *StatusError {
	return &StatusError{
		Code:    ErrCodeQueryFailed,
		Message: message,
	}
}

func IsQueryFailed(err error) bool {
	return isCode(err, ErrCodeQueryFailed)
}

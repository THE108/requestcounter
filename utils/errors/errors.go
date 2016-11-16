package errors

import (
	"fmt"
)

type CodedError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Wrap(err error, code int) *CodedError {
	return &CodedError{
		Code:    code,
		Message: err.Error(),
	}
}

func New(code int, msg string) *CodedError {
	return &CodedError{
		Code:    code,
		Message: msg,
	}
}

func Newf(code int, format string, args ...interface{}) *CodedError {
	return &CodedError{
		Code:    code,
		Message: fmt.Sprintf(format, args),
	}
}

func (ce *CodedError) String() string {
	return ce.Message
}

func (ce *CodedError) Error() string {
	return ce.Message
}

func (ce *CodedError) GetHttpCode() int {
	return ce.Code
}

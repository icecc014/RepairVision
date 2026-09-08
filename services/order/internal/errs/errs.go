package errs

import "fmt"

type Error struct {
	HTTPStatus int
	Message    string
}

func (e *Error) Error() string {
	return e.Message
}

func New(httpStatus int, message string) *Error {
	return &Error{HTTPStatus: httpStatus, Message: message}
}

func BadRequest(message string) *Error {
	return New(400, message)
}

func Unauthorized(message string) *Error {
	return New(401, message)
}

func Forbidden(message string) *Error {
	return New(403, message)
}

func NotFound(message string) *Error {
	return New(404, message)
}

func Conflict(message string) *Error {
	return New(409, message)
}

func Internal(err error) *Error {
	return New(500, fmt.Sprintf("系统繁忙：%v", err))
}

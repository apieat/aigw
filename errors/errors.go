package errors

import (
	"errors"
	"fmt"
)

type ResponseWithError struct {
	Error string
}

func (r *ResponseWithError) GetError() error {
	if r.Error == "" {
		return nil
	}
	return errors.New(r.Error)
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code, e.Message)
}

var ErrorInvalidResponse = &Error{
	Code:    2001,
	Message: "invalid response",
}

var ErrorEmptyPrompt = &Error{
	Code:    1001,
	Message: "empty prompt",
}

var ErrorNoFunctionCall = &Error{
	Code:    1002,
	Message: "no function call",
}

var ErrorTooManyRetry = &Error{
	Code:    1003,
	Message: "too many retry because of invalid response",
}

func NoFunctionCall(raw string) error {
	return &Error{
		Code:    1002,
		Message: fmt.Sprintf("no function call is responded in %s", raw),
	}
}

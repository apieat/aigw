package errors

import (
	"fmt"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
}

func (e *Error) RespondAsJson() bool {
	return true
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d]%s", e.Code, e.Message)
}

const (
	InvalidResponse = 2001
)

func ErrorInvalidResponse(reason string) *Error {
	return &Error{
		Code:    InvalidResponse,
		Message: "invalid response",
		Reason:  reason,
	}
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

var ErrorStreamingNotSupported = &Error{
	Code:    1004,
	Message: "streaming is not supported",
}

func NoFunctionCall(raw string) error {
	return &Error{
		Code:    1002,
		Message: fmt.Sprintf("no function call is responded in %s", raw),
	}
}

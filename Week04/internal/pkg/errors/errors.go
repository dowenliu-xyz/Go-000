package errors

import (
	"fmt"
)

var _ error = &StatusError{}

type StatusError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("error: code = %d desc = %s", e.Code, e.Message)
}

package http

import (
	xerrors "github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/errors"
	"net/http"
)

var errMapping = map[int]int{
	xerrors.ErrCodeSaveFailed: http.StatusInternalServerError,
}

func Status(err error) int {
	if code, ok := errMapping[xerrors.Code(err)]; ok {
		return code
	} else {
		return http.StatusInternalServerError
	}
}

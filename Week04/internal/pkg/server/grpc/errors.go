package grpc

import (
	xerrors "github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/errors"
	"google.golang.org/grpc/codes"
)

var errMapping = map[int]codes.Code{
	xerrors.ErrCodeUnknown: codes.Unknown,
}

func Code(err error) codes.Code {
	if code, ok := errMapping[xerrors.Code(err)]; ok {
		return code
	} else {
		return codes.Unknown
	}
}

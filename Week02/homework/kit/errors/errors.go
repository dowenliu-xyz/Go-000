package errors

import (
	"errors"
	"fmt"
)

// errResourceNotFound 表示未找到指定的资源（如数据库记录、文件等）
// 不导出，防止API表面积扩大。使用 IsErrResourceNotFound 判断
type errResourceNotFound struct {
	cause error
}

func (e *errResourceNotFound) Cause() error {
	return e.cause
}

func (e *errResourceNotFound) Error() string {
	return fmt.Sprintf("Resource not found: %v", e.cause)
}

func IsErrResourceNotFound(err error) bool {
	var target *errResourceNotFound
	return errors.As(err, &target)
}

func NewErrResourceNotFound(cause error) error {
	return &errResourceNotFound{cause}
}

package main

import (
	"errors"
	"fmt"
)

// Create a named type for our new error type.
type errorString string

// Implement the error interface.

func (e errorString) Error() string {
	return string(e)
}

// New creates interface values of type error.
func New(text string) error {
	return errorString(text)
}

type errorStruct struct {
	s string
}

func (e errorStruct) Error() string {
	return e.s
}

func NewError(text string) error {
	return errorStruct{text}
}

var ErrNamedType = New("EOF")
var ErrStructType = NewError("EOF")
var StdErrorType = errors.New("EOF")

// 演示为什么 errors.New() 返回 *errors.errorString ，而不是 errors.errorString
func main() {
	if ErrNamedType == New("EOF") {
		// True。因为潜在类型 string 值相等。可以轻易被创建同样的值
		fmt.Println("Named Type Error") // 输出
	}

	if ErrStructType == NewError("EOF") {
		// errorStruct 结构体是值类型，只要其 s 相等，整个值也就相等。可以轻易创建同样的值
		fmt.Println("Error: ", ErrStructType) // 输出
	}

	if StdErrorType == errors.New("EOF") {
		// False。因为 errors.New() 返回的类型为 *errors.errorString，其值是地址。
		// 尽管地址指向的 errors.errorString 的值是相等的，但并不是同一份数据，地址是不同的。
		// 可以创建出同样的 errors.errorString 值但不可能生成同样的地址。
		fmt.Println("Struct Type Error") // 不会输出
	}
}

package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("start a goroutine...")
	go func() {
		fmt.Println("hello")
		panic("oops!")
	}()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("where is my goroutine?") // 不会打印，因为野生 goroutine panic 导致进程退出
}

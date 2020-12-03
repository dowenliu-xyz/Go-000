package main

import (
	"fmt"
	"github.com/dowenliu-xyz/Go-000/Week02/internal/syncx"
	"time"
)

func main() {
	fmt.Println("start a goroutine...")
	syncx.Go(func() {
		fmt.Println("hello")
		panic("oops!")
	})
	time.Sleep(100 * time.Millisecond)
	fmt.Println("I'v see my goroutine gone...") // 可以打印出来
}

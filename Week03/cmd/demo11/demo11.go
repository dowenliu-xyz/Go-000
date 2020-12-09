package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup
var counter int = 0

// build with flag `-race`

func main() {
	for routine := 0; routine < 2; routine++ {
		wg.Add(1)
		go Routine(routine + 1)
	}

	wg.Wait()
	fmt.Printf("Final Counter: %d\n", counter)
}

func Routine(id int) {
	for count := 0; count < 2; count++ {
		value := counter
		time.Sleep(1 * time.Nanosecond) // 主动触发 goroutine 切换
		value++
		counter = value
	}

	wg.Done()
}

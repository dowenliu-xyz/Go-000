package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	done := make(chan struct{})
	var mu sync.Mutex
	count := 0

	// goroutine 1
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				mu.Lock()
				time.Sleep(100 * time.Millisecond)
				count++
				mu.Unlock()
			}
		}
	}()

	// goroutine 2
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		mu.Unlock()
	}
	close(done)
	log.Printf("g1 locks: %d\n", count)
}

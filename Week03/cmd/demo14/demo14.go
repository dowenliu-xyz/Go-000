package main

import (
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// NOT executable. Just for hinting.

func loadConfig() *map[string]interface{} {
	return &map[string]interface{}{
		"now": time.Now(),
	}
}

func requests() []string {
	return []string{"a", "b", "c"}
}

func main() {
	var config atomic.Value // holds current server configuration
	// Create initial config value and store into config.
	config.Store(loadConfig())
	go func() {
		// Reload config every 10 seconds
		// and update config value with the new version.
		for {
			time.Sleep(1 * time.Second)
			config.Store(loadConfig())
		}
	}()
	// Create worker goroutine that handle incoming requests
	// using the latest config value.
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		source := rand.NewSource(int64(i))
		ra := rand.New(source)
		wg.Add(1)
		go func() {
			for r := range requests() {
				time.Sleep(time.Duration(ra.Intn(3)) * time.Second)
				c := config.Load().(*map[string]interface{})
				// Handle request r using config c.
				log.Printf("handling request %v using config loaded at %s\n", r, (*c)["now"])
				_, _ = r, c
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

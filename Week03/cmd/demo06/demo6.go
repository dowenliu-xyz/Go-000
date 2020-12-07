package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// search simulates a function that finds a record based
// on a search term. It takes 200ms to perform this work.
func search(term string) (string, error) {
	time.Sleep(200 * time.Millisecond)
	return "some value", nil
}

// process is the work for the program. It finds a record
// then prints it.
func process(term string) error {
	record, err := search(term)
	if err != nil {
		return err
	}
	fmt.Println("Received:", record)
	return nil
}

// result wraps the return values from search. It allows us
// to pass both values across a single channel
type result struct {
	record string
	err    error
}

// processWithTimeout is the work for the program. It finds a record
// then prints it, It fails if it takes more than 100ms.
func processWithTimeout(term string) error {

	// Create a context that will be canceled in 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Make a channel for the goroutine to report its result.
	ch := make(chan result)

	// launch a goroutine to find the record. Create a result
	// from the returned values to send through the channel.
	go func() {
		record, err := search(term)
		ch <- result{record, err}
	}()

	// Block waiting to either receive from the goroutine's
	// channel or for the context to be canceled
	select {
	case <-ctx.Done():
		return errors.New("search canceled")
	case result := <-ch:
		if result.err != nil {
			return result.err
		}
		fmt.Println("Received:", result.record)
		return nil
	}
}

func main() {
	err := process("without-timeout")
	if err != nil {
		fmt.Printf("process: error: %v\n", err)
	}
	err = processWithTimeout("with-timeout")
	if err != nil {
		fmt.Printf("processWithTimeout: error: %v\n", err)
	}
}

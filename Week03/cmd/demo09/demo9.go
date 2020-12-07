package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// Tracker knows how to track events for the application.
type Tracker struct {
	wg sync.WaitGroup
}

// Event starts tracking an event. It runs asynchronously to
// not block the caller. Be sure to call the Shutdown function
// before the program exits so all tracked events finish.
func (t *Tracker) Event(data string) {

	// Increment counter so Shutdown knows to wait for this event.
	t.wg.Add(1)

	// Track event in a goroutine so caller is not blocked.
	go func() {
		// Decrement counter to tell Shutdown this goroutine finished.
		defer t.wg.Done()

		time.Sleep(time.Millisecond) // Simulate network write latency.
		log.Println(data)
	}()
}

// Shutdown waits for all tracked events to finish processing
// or for the provided context to be canceled.
func (t *Tracker) Shutdown(ctx context.Context) error {

	// Create a channel to signal when the waitgroup is finished.
	ch := make(chan struct{})

	// Create a goroutine to wait for all other goroutines to
	// be done then close the channel to unblock the select.
	go func() {
		t.wg.Wait()
		close(ch)
	}()

	// Block this function from returning. Wait for either the
	// waitgroup to finish or the context to expire.
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// App holds application state.
type App struct {
	track Tracker
}

// Handle represents an example handler for the web service
func (a *App) Handle(w http.ResponseWriter, r *http.Request) {
	// Do some actual work.

	// Respond to the client.
	w.WriteHeader(http.StatusCreated)

	// Track the event.
	a.track.Event("this event")
}

func main() {

	// Start a server.
	// Details not shown...
	var a App

	// Shut the server down.
	// Details not shown...

	// Wait up to 5 seconds for all event goroutines to finish.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := a.track.Shutdown(ctx)
	if err != nil {
		log.Println(err)
	}

}

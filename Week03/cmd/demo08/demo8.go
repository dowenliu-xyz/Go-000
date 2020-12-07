package main

import (
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

// Shutdown waits for all tracked events to finish processing.
func (t *Tracker) Shutdown() {
	t.wg.Wait()
}

// App holds application state.
type App struct {
	track Tracker
}

// Handle represents an example handler for the web service.
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

	// Wait for all event goroutines to finish
	a.track.Shutdown()
}

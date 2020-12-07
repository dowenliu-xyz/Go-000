package main

import (
	"log"
	"net/http"
	"time"
)

// Tracker knows how to track events for the application.
type Tracker struct {
}

// Event records an event to a database or stream.
func (t *Tracker) Event(data string) {
	time.Sleep(time.Millisecond) // Simulate network write latency.
	log.Println(data)
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

	// Fire and Hope.
	// BUG: We are not managing this goroutine.
	go a.track.Event("this event")
}

func main() {
	// Start a server.
	// Details not shown...
	var a App

	// Shut the server down.
	// Details not shown...
	_ = a
}

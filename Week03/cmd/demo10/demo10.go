package main

import (
	"context"
	"fmt"
	"time"
)

// Tracker knows how to track events for the application.
type Tracker struct {
	ch   chan string
	stop chan struct{}
}

func (t *Tracker) Events(ctx context.Context, data string) error {
	select { // 如果阻塞在 t.ch <- 且 ctx 又不取消，这时 Server shutdown，数据丢失。
	case t.ch <- data: // ch 满或 closed 会阻塞
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Tracker) Run() {
	for data := range t.ch {
		time.Sleep(1 * time.Second)
		fmt.Println(data)
	}
	t.stop <- struct{}{}
}

func (t *Tracker) Shutdown(ctx context.Context) {
	close(t.ch)
	select {
	case <-t.stop:
	case <-ctx.Done():
	}
}

func NewTracker() *Tracker {
	return &Tracker{
		ch: make(chan string, 10),
	}
}

func main() {
	tr := NewTracker()
	go tr.Run()
	// 如果这样的 track 并行有10个以上，不会阻塞吗？
	_ = tr.Events(context.Background(), "test")
	_ = tr.Events(context.Background(), "test")
	_ = tr.Events(context.Background(), "test") // 如果这样的 track 并行有10个以上，不会阻塞吗？
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	defer cancel()
	tr.Shutdown(ctx)
}

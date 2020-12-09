package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func (c *Config) T() {

}

func BenchmarkAtomic(b *testing.B) {
	var v atomic.Value
	v.Store(&Config{})
	stop := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				return
			default:
			}
			i++
			cfg := &Config{a: []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}}
			v.Store(cfg)
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for n := 0; n < 4; n++ {
			wg.Add(1)
			go func() {
				for n := 0; n < 100; n++ {
					cfg := v.Load().(*Config)
					cfg.T()
					//fmt.Printf("%v\n", cfg)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
	b.StopTimer()
	close(stop)
}

func BenchmarkMutex(b *testing.B) {
	cfg := &Config{}
	var mu sync.Mutex
	stop := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				return
			default:
			}
			i++
			mu.Lock()
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
			mu.Unlock()
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for n := 0; n < 4; n++ {
			wg.Add(1)
			go func() {
				for n := 0; n < 100; n++ {
					mu.Lock()
					cfg.T()
					mu.Unlock()
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
	b.StopTimer()
	close(stop)
}

func BenchmarkRWMutex(b *testing.B) {
	cfg := &Config{}
	var mu sync.RWMutex
	stop := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-stop:
				return
			default:
			}
			i++
			mu.Lock()
			cfg.a = []int{i, i + 1, i + 2, i + 3, i + 4, i + 5}
			mu.Unlock()
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for n := 0; n < 4; n++ {
			wg.Add(1)
			go func() {
				for n := 0; n < 100; n++ {
					mu.RLock()
					cfg.T()
					mu.RUnlock()
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
	b.StopTimer()
	close(stop)
}

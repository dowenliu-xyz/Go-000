package main

import (
	"context"
	"fmt"
	"net/http"
)

func serve(addr string, handler http.Handler, stop <-chan struct{}) error {
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}
	go func() {
		<-stop // wait for stop signal
		s.Shutdown(context.Background())
	}()
	return s.ListenAndServe()
}

func serveApp(stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "Hello, QCon!")
	})
	return serve("0.0.0.0:8080", mux, stop)
}

func serveDebug(stop <-chan struct{}) error {
	return serve("127.0.0.1:8081", http.DefaultServeMux, stop)
}

func main() {
	done := make(chan error, 2)
	stop := make(chan struct{})
	go func() {
		done <- serveDebug(stop)
	}()
	go func() {
		done <- serveApp(stop)
	}()

	// serveApp、serveDebug 任意一个出错，都会解除 <-done的阻塞
	// close(stop) 会广播解除所有 <-stop 的阻塞，没有出错的监听也会被 shutdown
	var stopped bool
	for i := 0; i < cap(done); i++ { // 循环两次是为了等所有的 server 平滑安全退出
		if err := <-done; err != nil {
			fmt.Printf("error: %v\n", err)
		}
		if !stopped {
			stopped = true
			close(stop)
		}
	}

}

package main

import (
	"fmt"
	"net/http"
)

func serveApp() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "Hello, QCon!")
	})
	http.ListenAndServe("0.0.0.0:8080", mux)
}

func serveDebug() {
	http.ListenAndServe("127.0.0.1:8001", http.DefaultServeMux)
}

func main() {
	// 将逻辑分解为各自的函数，使 main 流程简洁，但还是没有解决 demo02 的问题
	go serveDebug()
	serveApp()
}

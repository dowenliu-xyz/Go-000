package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(resp, "Hello, QCon!")
	})
	// ↓ 如果不使用 go，会阻塞在这一行，再下一行的 ListenAndServe 就没有机会执行
	//   但这里启动后不管了，这种做法是不好的，应该管理 goroutine 的结束。启动者要对 goroutine 的生命周期负责。
	go http.ListenAndServe("127.0.0.1:0801", http.DefaultServeMux)
	http.ListenAndServe("0.0.0.0:8080", mux)
}

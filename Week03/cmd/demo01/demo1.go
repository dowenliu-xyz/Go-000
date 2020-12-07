package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, GopherCon SG")
	})
	go func() {
		// 是否会出错，main goroutine 感知不到，也处理不了。
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err) // Fatal() 底层调用了 os.Exit()
		}
	}()
	select { // 永远阻塞。
	}
}

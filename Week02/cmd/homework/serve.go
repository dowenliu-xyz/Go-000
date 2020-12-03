package main

import (
	"encoding/json"
	kithttp "github.com/dowenliu-xyz/Go-000/Week02/homework/kit/http"
	"github.com/dowenliu-xyz/Go-000/Week02/homework/staff/endpoint"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

func Chain(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				err1, ok := err.(error)
				if ok {
					type stackTracer interface {
						StackTrace() errors.StackTrace
					}
					err2, ok := err1.(stackTracer)
					if ok {
						log.Printf("err recoved: %+v", err2)
					} else {
						log.Printf("err recoved: %+v", errors.WithStack(err1))
					}
				} else {
					log.Printf("err recoved: %+v", errors.Errorf("%v", err))
				}
				w.WriteHeader(500)
				status := kithttp.Status{Code: 500, Message: "服务器内部错误"}
				bytes, err := json.Marshal(status)
				if err != nil {
					log.Printf("Marshal 结果响应体失败 %v", err)
					return
				}
				_, _ = w.Write(bytes)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func main() {
	router := httprouter.New()
	router.GET("/staff/:id", endpoint.GetStaff)
	err := http.ListenAndServe(":8080", Chain(router))
	if err != nil {
		log.Fatal(err)
	}
}

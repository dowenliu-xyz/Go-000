package http

import "net/http"

type Middleware func(handler http.Handler) http.Handler

func Chain(handle http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handle = middleware(handle)
	}
	return handle
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	kithttp "github.com/dowenliu-xyz/Go-000/Week03/internal/kit/http"
	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var appAddr string
var debugAddr string

func init() {
	flag.StringVar(&appAddr, "appAddr", ":8080", "Specify app serving address")
	flag.StringVar(&debugAddr, "debugAddr", "127.0.0.1:8081", "Specify debug serving address")
}

func main() {
	// 解析应用 flag 参数
	flag.Parse()
	log.Printf("Pid: %d\n", os.Getpid())

	// 构造服务组合
	server := NewCompositeServer(context.Background())

	// 注册信号量处理
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // 关注 Ctrl-C 和 kill -1
	go func() {
		sig := <-sigs
		log.Printf("Recieved signal: %v\n", sig)
		server.Shutdown()
	}()
	// 监听并提供服务
	server.ListenAndServe()
}

// CompositeServer 组合了应用服务和调试服务，同时保持了服务 context 及其 cancel 方法
type CompositeServer struct {
	appServer   *http.Server
	debugServer *http.Server
	serveCtx    context.Context
	cancel      context.CancelFunc
	done        chan struct{}
}

// 构造 CompositeServer。传入一个上下文
func NewCompositeServer(ctx context.Context) *CompositeServer {
	serveCtx, cancelFunc := context.WithCancel(ctx)
	return &CompositeServer{
		appServer:   buildAppServer(),
		debugServer: buildDebugServer(cancelFunc),
		serveCtx:    serveCtx,
		cancel:      cancelFunc,
		done:        make(chan struct{}, 2),
	}
}

// 监听并启动
func (s *CompositeServer) ListenAndServe() {
	serveGroup := errgroup.WithCancel(s.serveCtx)
	// 启动 servers
	serveGroup.Go(func(ctx context.Context) error {
		log.Printf("Starting app server at %s\n", s.appServer.Addr)
		return s.appServer.ListenAndServe()
	})
	serveGroup.Go(func(ctx context.Context) error {
		log.Printf("Starting debug server at %s\n", s.debugServer.Addr)
		return s.debugServer.ListenAndServe()
	})
	// 保证平滑关闭
	serveGroup.Go(func(ctx context.Context) error {
		// 等待主动关闭或出错关闭
		<-ctx.Done()
		log.Printf("Shutting down app server...")
		timeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := s.appServer.Shutdown(timeout)
		if err != nil {
			log.Printf("App server stopped: %+v\n", err)
		} else {
			log.Println("App server stopped.")
		}
		s.done <- struct{}{}
		return err
	})
	serveGroup.Go(func(ctx context.Context) error {
		// 等待主动关闭或出错关闭
		<-ctx.Done()
		log.Printf("Shutting down debug server...")
		timeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := s.debugServer.Shutdown(timeout)
		if err != nil {
			log.Printf("Debug server stopped: %+v\n", err)
		} else {
			log.Println("Debug server stopped.")
		}
		s.done <- struct{}{}
		return err
	})

	// 等待任一服务停止或出错，context 取消
	err := serveGroup.Wait()
	if err != nil {
		log.Printf("Shutdown: %+v\n", err)
	} else {
		log.Println("Shutdown.")
	}

	// 等待所有服务正常关闭
	for i := 0; i < cap(s.done); i++ {
		<-s.done
	}
	log.Printf("All servers stopped.")
}

func (s *CompositeServer) Shutdown() {
	s.cancel()
}

// 构建 appServer
func buildAppServer() *http.Server {
	router := httprouter.New()
	router.GET("/", func(writer http.ResponseWriter, _ *http.Request, params httprouter.Params) {
		name := params.ByName("name")
		if name == "" {
			name = "World"
		}
		writer.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		writer.WriteHeader(200)
		_, err := fmt.Fprintf(writer, "Hello, %s!", name)
		if err != nil {
			log.Printf("write response err: %v", err)
		}
	})
	return &http.Server{
		Addr:    appAddr,
		Handler: Chain(router),
	}
}

// 构建 debugServer。cancelFunc 用于通知结束服务
func buildDebugServer(cancelFunc context.CancelFunc) *http.Server {
	http.HandleFunc("/shutdown", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/plain;charset=UTF-8")
		writer.WriteHeader(200)
		_, err := fmt.Fprintln(writer, "Shutting down...")
		if err != nil {
			log.Printf("write response err: %v", err)
		}
		log.Println("Shutdown request received, shutting down...")
		cancelFunc()
	})
	return &http.Server{
		Addr:    debugAddr,
		Handler: Chain(http.DefaultServeMux),
	}
}

// Chain 构造 http.Handler 链。进行 panic recover() 兜底
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

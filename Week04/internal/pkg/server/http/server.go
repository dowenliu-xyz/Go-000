package http

import (
	"bytes"
	"context"
	"fmt"
	xgrpc "github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/server/grpc"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"net/http"
	"strings"
)

type ServiceWithName func(cb func(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error), interceptors ...grpc.UnaryServerInterceptor) (string, string, http.HandlerFunc)

func restPath(service, method string, hf http.HandlerFunc) (string, http.HandlerFunc) {
	return fmt.Sprintf("/%s/%s", strings.ToLower(service), strings.ToLower(method)), hf
}

func logCallback(ctx context.Context, w http.ResponseWriter, r *http.Request, arg, ret proto.Message, err error) {
	log.Printf("INFO: call %s: arg: {%v}, ret: {%s}", r.RequestURI, arg, ret)
	if err == nil {
		return
	}
	log.Printf("ERROR: %+v", err)
	w.WriteHeader(Status(err))
	p := status.New(xgrpc.Code(err), err.Error()).Proto()
	switch r.Header.Get("Content-Type") {
	case "application/protobuf", "application/x-protobuf":
		buf, err := proto.Marshal(p)
		if err != nil {
			return
		}
		if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
			return
		}
	case "application/json":
		buf, err := protojson.Marshal(p)
		if err != nil {
			return
		}
		if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
			return
		}
	default:
	}
}

type Server struct {
	*http.Server
	mux *http.ServeMux
}

func NewServer(addr string, middlewares ...Middleware) *Server {
	var handler http.Handler
	mux := http.NewServeMux()
	handler = mux
	s := &Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: Chain(handler, middlewares...),
		},
		mux: mux,
	}
	return s
}

func (s *Server) HandleServiceWithName(swn ServiceWithName) {
	s.mux.Handle(restPath(swn(logCallback)))
}

func (s *Server) Handle(path string, handler http.Handler) {
	s.mux.Handle(path, handler)
}

func (s *Server) HandleFunc(path string, h func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(path, h)
}

func (s *Server) Start(_ context.Context) error {
	return s.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.Shutdown(ctx); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

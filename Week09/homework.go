package main

import (
	"bufio"
	"context"
	"flag"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var addr string

func init() {
	flag.StringVar(&addr, "addr", "0.0.0.0:7777", "TCP server listen address")
}

func main() {
	flag.Parse()
	server := NewServer(addr)
	ctx, cancel := context.WithCancel(context.Background())
	err := server.Start(ctx)
	if err != nil {
		log.Printf("Error starting server, %+v", err)
		return
	}
	log.Printf("server started, listening at %s", addr)
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	select {
	case <-sign:
		cancel()
	}
	log.Printf("server stopping")
}

type Server struct {
	addr string
	l    net.Listener
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Start(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		return errors.Wrapf(err, "listen to %s failed", s.addr)
	}
	s.l = listen
	go func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
			}
		}()
		s.listen(ctx)
	}(ctx)
	return nil
}

func (s *Server) listen(ctx context.Context) {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			log.Printf("accept error: %+v", err)
		} else {
			handler := &ConnHandler{
				conn: conn,
			}
			handler.ch = make(chan string, 1)
			handler.handle(ctx)
		}
		select {
		case <-ctx.Done():
			log.Printf("stop listening...")
			return
		default:
			continue
		}
	}
}

type ConnHandler struct {
	conn    net.Conn
	stopped bool
	mut     sync.Mutex
	ch      chan string
}

func (handler *ConnHandler) handle(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	go func(ctx context.Context) {
		reader := bufio.NewReader(handler.conn)
		for {
			line, _, err := reader.ReadLine()
			if err != nil {
				log.Printf("read error: %+v", err)
				cancel()
				return
			}
			msg := string(line)
			log.Printf("read: %s", msg)
			handler.ch <- msg
			select {
			case <-ctx.Done():
				handler.stop()
				log.Printf("stop reading from client")
				return
			default:
				continue
			}
		}
	}(ctx)
	go func(ctx context.Context) {
		bw := bufio.NewWriter(handler.conn)
		writer := &ErrWriter{
			bw,
			nil,
		}
		for {
			select {
			case msg := <-handler.ch:
				if msg == "Bye" {
					log.Printf("client said Goodbye, say goodbye to it.")
					writer.WriteString("Goodbye!\n").Flush()
					if writer.err != nil {
						log.Printf("error writing/flushing, %+v", writer.err)
					}
					cancel()
					return
				}
				writer.WriteString("received: ").WriteString(msg).WriteString("\n")
				writer.Flush()
				if writer.err != nil {
					log.Printf("error writing/flushing, %+v", writer.err)
					cancel()
					return
				}
			case <-ctx.Done():
				handler.stop()
				log.Printf("stop writing to client")
				return
			}
		}
	}(ctx)
}

func (handler *ConnHandler) stop() {
	handler.mut.Lock()
	defer handler.mut.Unlock()
	if !handler.stopped {
		_ = handler.conn.Close()
		handler.stopped = true
	}
}

type ErrWriter struct {
	*bufio.Writer
	err error
}

func (w *ErrWriter) WriteString(s string) *ErrWriter {
	if w.err != nil {
		return w
	}
	_, w.err = w.Writer.WriteString(s)
	return w
}

func (w *ErrWriter) Flush() *ErrWriter {
	if w.err != nil {
		return w
	}
	w.err = w.Writer.Flush()
	return w
}

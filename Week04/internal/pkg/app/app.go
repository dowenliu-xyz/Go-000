package app

import (
	"context"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 抄 kratos 作业。。。

type Hook struct {
	OnStart func(ctx context.Context) error
	OnStop  func(ctx context.Context) error
}

type options struct {
	startTimeout time.Duration
	stopTimeout  time.Duration

	sigs  []os.Signal
	sigFn func(*App, os.Signal)
}

type Option func(o *options)

func StartTimeout(d time.Duration) Option {
	return func(o *options) { o.startTimeout = d }
}

func StopTimeout(d time.Duration) Option {
	return func(o *options) { o.stopTimeout = d }
}

func Signal(fn func(*App, os.Signal), sigs ...os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
		o.sigFn = fn
	}
}

type App struct {
	options options
	hooks   []Hook

	cancel func()
}

func New(opts ...Option) *App {
	options := options{
		startTimeout: time.Second * 30,
		stopTimeout:  time.Second * 30,
		sigs: []os.Signal{
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGINT,
		},
		sigFn: func(a *App, sig os.Signal) {
			switch sig {
			case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM:
				a.Stop()
			default:
			}
		},
	}
	for _, opt := range opts {
		opt(&options)
	}
	return &App{options: options}
}

func (a *App) Append(hook Hook) {
	a.hooks = append(a.hooks, hook)
}

func (a *App) Run() error {
	var ctx context.Context
	ctx, a.cancel = context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	for _, hook := range a.hooks {
		hook := hook
		if hook.OnStop != nil {
			g.Go(func() error {
				<-ctx.Done() // wait for stop signal
				stopCtx, cancel := context.WithTimeout(context.Background(), a.options.stopTimeout)
				defer cancel()
				return hook.OnStop(stopCtx)
			})
		}
		if hook.OnStart != nil {
			g.Go(func() error {
				startCtx, cancel := context.WithTimeout(context.Background(), a.options.startTimeout)
				defer cancel()
				return hook.OnStart(startCtx)
			})
		}
	}
	if len(a.options.sigs) == 0 {
		return g.Wait()
	}
	c := make(chan os.Signal, len(a.options.sigs))
	signal.Notify(c, a.options.sigs...)
	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-c:
				if a.options.sigFn != nil {
					a.options.sigFn(a, sig)
				}
			}
		}
	})
	return g.Wait()
}

func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
}

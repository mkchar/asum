package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Engine struct {
	app  *fiber.App
	cfg  Config
	addr string
}

func New(opts ...Option) *Engine {
	o := Options{
		Config: DefaultConfig(),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	conf := fiber.Config{
		AppName:            o.Config.AppName,
		ReadTimeout:        o.Config.ReadTimeout,
		WriteTimeout:       o.Config.WriteTimeout,
		IdleTimeout:        o.Config.IdleTimeout,
		BodyLimit:          o.Config.BodyLimit,
		Concurrency:        o.Config.Concurrency,
		CaseSensitive:      o.Config.CaseSensitive,
		StrictRouting:      o.Config.StrictRouting,
		Immutable:          o.Config.Immutable,
		UnescapePath:       o.Config.UnescapePath,
		DisableKeepalive:   o.Config.DisableKeepalive,
		DisableDefaultDate: o.Config.DisableDefaultDate,
		ServerHeader:       o.Config.ServerHeader,
	}

	app := fiber.New(conf)

	return &Engine{
		app:  app,
		cfg:  o.Config,
		addr: fmt.Sprintf("%s:%d", o.Config.Host, o.Config.Port),
	}
}

func (e *Engine) App() *fiber.App {
	return e.app
}

func (e *Engine) Config() Config {
	return e.cfg
}

func (e *Engine) Addr() string {
	return e.addr
}

func (e *Engine) Listen() error {
	return e.app.Listen(e.addr)
}

func (e *Engine) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- e.app.Listen(e.addr)
	}()

	select {
	case <-ctx.Done():
		_ = e.Shutdown(5 * time.Second)
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (e *Engine) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.app.ShutdownWithContext(ctx)
}

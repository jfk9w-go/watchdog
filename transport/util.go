package transport

import (
	"context"
	"io"

	"github.com/jfk9w-go/flu"
)

type ContextFunc func(context.Context) (context.Context, func())

type Dispatcher struct {
	Cancel func()
	Closer io.Closer
	flu.WaitGroup
}

func (d *Dispatcher) GoWithCancel(ctx context.Context, rateLimiter flu.RateLimiter, fun func(context.Context)) {
	d.Cancel = d.WaitGroup.Go(ctx, rateLimiter, fun)
}

func (d *Dispatcher) Close() error {
	d.Cancel()
	err := d.Closer.Close()
	d.Wait()
	return err
}

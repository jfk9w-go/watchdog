package client

import (
	"context"
	"time"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/flu/serde"
	"github.com/pkg/errors"
)

type Config struct {
	ServiceID string         `yaml:"service_id"`
	Hostname  string         `yaml:"hostname"`
	Interval  serde.Duration `yaml:"interval"`
	Address   string         `yaml:"address"`
}

type Instance struct {
	id     string
	start  time.Time
	client Wrapper
	cancel func()
	flu.WaitGroup
}

func Run(ctx context.Context, clock flu.Clock, config Config, transport Transport) *Instance {
	client, err := transport.NewClient(ctx)
	if err != nil {
		panic(errors.Wrap(err, "create client"))
	}

	i := &Instance{
		client: Wrapper{
			Service:   client,
			ServiceID: config.ServiceID,
			Hostname:  config.Hostname,
			Clock:     clock,
		},
	}

	i.start = clock.Now()
	i.cancel = i.Go(ctx, nil, func(ctx context.Context) {
		ticker := time.NewTicker(config.Interval.Duration)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				i.client.Send(ctx, i.start, Active, nil)
			case <-ctx.Done():
				return
			}
		}
	})

	return i
}

func (i *Instance) Complete(ctx context.Context) {
	err := recover()
	i.cancel()
	i.Wait()
	i.client.Send(ctx, i.start, Complete, err)
	_ = i.client.Close()
}

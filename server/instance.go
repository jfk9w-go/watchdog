package server

import (
	"context"
	"io"

	"github.com/jfk9w/watchdog/client"

	"github.com/jfk9w/watchdog/transport"
)

type Wrapper struct {
	client.Service
}

type Instance struct {
	client.Service
	listeners []io.Closer
}

func (i *Instance) Run(ctx context.Context, address string) error {
	return i.RunAdvanced(ctx, transport.New(address))
}

func (i *Instance) RunAdvanced(ctx context.Context, transport client.Transport) error {
	listener, err := transport.RunListener(ctx, Wrapper{Service: i.Service})
	if err != nil {
		return err
	}

	i.addListener(listener)
	return nil
}

func (i *Instance) addListener(listener io.Closer) {
	if i.listeners == nil {
		i.listeners = []io.Closer{listener}
	} else {
		i.listeners = append(i.listeners, listener)
	}
}

func (i *Instance) Close() error {
	for _, listener := range i.listeners {
		_ = listener.Close()
	}

	return i.Service.Close()
}

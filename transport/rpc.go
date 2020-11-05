package transport

import (
	"context"
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/jfk9w-go/watchdog/client"

	"github.com/jfk9w-go/flu"
	"github.com/pkg/errors"
)

type RPC struct {
	Network     string
	Address     string
	ContextFunc ContextFunc
	Codec       interface {
		Client(conn io.ReadWriteCloser) rpc.ClientCodec
		Server(conn io.ReadWriteCloser) rpc.ServerCodec
	}
}

func (r *RPC) WithDefaultCodec() *RPC {
	r.Codec = JSONRPC{}
	return r
}

func (r *RPC) NewClient(ctx context.Context) (client.Service, error) {
	conn, err := flu.Conn{
		Context: ctx,
		Network: r.Network,
		Address: r.Address,
	}.Dial()
	if err != nil {
		return nil, errors.Wrap(err, "dial")
	}

	return &rpcClient{Client: rpc.NewClientWithCodec(r.Codec.Client(conn))}, nil
}

func (r *RPC) RunListener(ctx context.Context, service client.Service) (io.Closer, error) {
	listener, err := new(net.ListenConfig).Listen(ctx, r.Network, r.Address)
	if err != nil {
		return nil, errors.Wrap(err, "listen")
	}

	server := rpc.NewServer()
	if err := server.RegisterName(RPCName, &RPCWrapper{
		ctx:     ctx,
		ctxFunc: r.ContextFunc,
		service: service,
	}); err != nil {
		return nil, errors.Wrap(err, "register wrapper")
	}

	dispatcher := &Dispatcher{Closer: listener}
	dispatcher.GoWithCancel(ctx, flu.RateUnlimiter, func(ctx context.Context) {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if ctx.Err() != nil || errors.Cause(err) == rpc.ErrShutdown {
					return
				} else {
					log.Printf("[watchdog] failed to accept connection: %v", err)
					continue
				}
			}

			dispatcher.Go(ctx, nil, func(_ context.Context) {
				codec := r.Codec.Server(conn)
				server.ServeCodec(codec)
			})
		}
	})

	return dispatcher, nil
}

const RPCName = "Watchdog"

type JSONRPC struct {
}

func (JSONRPC) Client(conn io.ReadWriteCloser) rpc.ClientCodec {
	return jsonrpc.NewClientCodec(conn)
}

func (JSONRPC) Server(conn io.ReadWriteCloser) rpc.ServerCodec {
	return jsonrpc.NewServerCodec(conn)
}

type rpcClient struct {
	*rpc.Client
}

func (c *rpcClient) Alert(_ context.Context, alert client.Alert) error {
	return c.Call(RPCName+".Alert", alert, new(client.Void))
}

type RPCWrapper struct {
	ctx     context.Context
	ctxFunc ContextFunc
	service client.Service
}

func (w *RPCWrapper) Alert(alert client.Alert, _ *client.Void) error {
	ctx := w.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var cancel func()
	if w.ctxFunc != nil {
		ctx, cancel = w.ctxFunc(ctx)
	}

	if cancel != nil {
		defer cancel()
	}

	return w.service.Alert(ctx, alert)
}

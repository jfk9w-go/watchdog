package transport

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/jfk9w-go/flu"
	fluhttp "github.com/jfk9w-go/flu/http"
	"github.com/jfk9w-go/watchdog/client"
	"github.com/pkg/errors"
)

func New(address string) *HTTP {
	return &HTTP{
		Address: address,
		ContextFunc: func(ctx context.Context) (context.Context, func()) {
			return context.WithTimeout(ctx, 1*time.Minute)
		},
		Codec: JSON,
	}
}

type BinaryCodec interface {
	flu.EncoderTo
	flu.DecoderFrom
}

type BinaryCodecFactory func(interface{}) BinaryCodec

type HTTP struct {
	Address     string
	ContextFunc ContextFunc
	Codec       BinaryCodecFactory
}

func JSON(value interface{}) BinaryCodec {
	return flu.JSON{Value: value}
}

func XML(value interface{}) BinaryCodec {
	return flu.XML{Value: value}
}

type httpClient struct {
	*fluhttp.Client
	address string
	codec   BinaryCodecFactory
}

func (h httpClient) Alert(ctx context.Context, alert client.Alert) error {
	return h.POST(h.address).
		BodyEncoder(h.codec(alert)).
		Context(ctx).
		Execute().
		CheckStatus(http.StatusOK).
		Error
}

func (h httpClient) Close() error {
	return nil
}

func (h *HTTP) NewClient(_ context.Context) (client.Service, error) {
	return httpClient{
		Client:  fluhttp.NewClient(nil),
		address: h.Address,
		codec:   h.Codec,
	}, nil
}

func (h *HTTP) RunListener(ctx context.Context, service client.Service) (io.Closer, error) {
	u, err := url.Parse(h.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "parse address: %s", h.Address)
	}

	ctx, cancel := context.WithCancel(ctx)
	dispatcher := &Dispatcher{Cancel: cancel}
	mux := http.NewServeMux()
	mux.Handle(u.Path, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		dispatcher.Add(1)
		defer dispatcher.Done()

		ctx, cancel := h.ContextFunc(ctx)
		defer cancel()

		var alert client.Alert
		if err := flu.DecodeFrom(flu.IO{R: request.Body}, h.Codec(&alert)); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Printf("[watchdog] failed to decode request: %s", err)
			return
		}

		if err := service.Alert(ctx, alert); err != nil {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			log.Printf("[watchdog] failed to process %v: %s", alert, err)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}))

	server := &http.Server{
		Addr:    u.Host,
		Handler: mux,
	}

	dispatcher.Closer = server
	go func() { log.Printf("[watchdog] server shutdown: %s", server.ListenAndServe()) }()
	return dispatcher, nil
}

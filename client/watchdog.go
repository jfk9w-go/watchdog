package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jfk9w-go/flu"
)

type State uint8

const (
	Active State = iota
	Complete
)

type Alert struct {
	ServiceID string    `json:"service_id"`
	Hostname  string    `json:"hostname"`
	Start     time.Time `json:"start"`
	Time      time.Time `json:"time"`
	State     State     `json:"state"`
	Error     string    `json:"error"`
}

type Void struct {
}

type Transport interface {
	NewClient(ctx context.Context) (Service, error)
	RunListener(ctx context.Context, watchdog Service) (io.Closer, error)
}

type Service interface {
	io.Closer
	Alert(context.Context, Alert) error
}

type Wrapper struct {
	Service
	ServiceID string
	Hostname  string
	Clock     flu.Clock
}

func (w Wrapper) Send(ctx context.Context, start time.Time, state State, err interface{}) {
	errDesc := ""
	if err != nil {
		errDesc = fmt.Sprintf("%+v", err)
	}

	if err := w.Alert(ctx, Alert{
		ServiceID: w.ServiceID,
		Hostname:  w.Hostname,
		Start:     start,
		Time:      w.Clock.Now(),
		State:     state,
		Error:     errDesc}); err != nil {
		log.Printf("[watchdog] alert error: %v", err)
	}
}

func Hostname(defaultValue string) string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("[watchdog] failed to get hostname: %s", err)
		return defaultValue
	}

	return hostname
}

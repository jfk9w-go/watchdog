package watchdog

import (
	"context"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/watchdog/client"
	"github.com/jfk9w-go/watchdog/transport"
)

type (
	Alert    = client.Alert
	Instance = client.Instance
	Config   = client.Config
)

func Run(ctx context.Context, config Config) *Instance {
	if config.Hostname == "" {
		config.Hostname = Hostname("unknown")
	}

	return client.Run(ctx, flu.DefaultClock, config, transport.New(config.Address))
}

var Hostname = client.Hostname

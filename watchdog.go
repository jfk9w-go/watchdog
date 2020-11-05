package watchdog

import (
	"context"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/watchdog/client"
	"github.com/jfk9w-go/watchdog/transport"
)

type (
	Alert   = client.Alert
	Service = client.Service
	Config  = client.Config
)

func Run(ctx context.Context, config client.Config) *client.Instance {
	if config.Hostname == "" {
		config.Hostname = Hostname("unknown")
	}

	return client.Run(ctx, flu.DefaultClock, config, transport.New(config.Address))
}

var Hostname = client.Hostname

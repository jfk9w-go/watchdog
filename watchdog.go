package watchdog

import (
	"context"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w/watchdog/client"
	"github.com/jfk9w/watchdog/transport"
)

type (
	Alert   = client.Alert
	Service = client.Service
)

func Run(ctx context.Context, config client.Config) *client.Instance {
	if config.Hostname == "" {
		config.Hostname = Hostname("unknown")
	}

	return client.Run(ctx, flu.DefaultClock, config, transport.New(config.Address))
}

var Hostname = client.Hostname

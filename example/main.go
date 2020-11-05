package main

import (
	"context"
	"time"

	"github.com/jfk9w-go/flu/serde"
	"github.com/jfk9w/watchdog"
	"github.com/jfk9w/watchdog/client"
	"github.com/jfk9w/watchdog/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := server.NewDatabase("postgres", "postgresql://hikkabot@192.168.88.4:5432/hikkabot_test")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Init(ctx); err != nil {
		panic(err)
	}

	s := &server.Instance{Service: db}
	defer s.Close()

	config := client.Config{
		Address:   "http://localhost:12345/",
		ServiceID: "test-service",
		Interval:  serde.Duration{Duration: time.Second},
	}

	if err := s.Run(ctx, config.Address); err != nil {
		panic(err)
	}

	defer watchdog.Run(ctx, config).Complete(ctx)
	time.Sleep(10*time.Second + 40*time.Millisecond)
	panic("kek")
}

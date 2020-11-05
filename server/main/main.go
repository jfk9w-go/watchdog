package main

import (
	"context"
	"os"
	"syscall"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/watchdog/server"
)

type Config struct {
	Address  string `yaml:"address"`
	Database struct {
		Driver     string `yaml:"driver"`
		Datasource string `yaml:"datasource"`
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var config Config
	if err := flu.DecodeFrom(flu.File(os.Args[1]), flu.JSON{Value: &config}); err != nil {
		panic(err)
	}

	db, err := server.NewDatabase(config.Database.Driver, config.Database.Datasource)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Init(ctx); err != nil {
		panic(err)
	}

	s := &server.Instance{Service: db}
	defer s.Close()

	if err := s.Run(ctx, config.Address); err != nil {
		panic(err)
	}
	flu.AwaitSignal(syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)
}

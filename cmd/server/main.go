package main

import (
	"context"
	"flag"

	serverApp "github.com/zvfkjytytw/humay/internal/server/app"
	humayHTTPServer "github.com/zvfkjytytw/humay/internal/server/http"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./build/server.yaml", "Server config file")
	flag.Parse()

	config := &serverApp.ServerConfig{
		HTTPConfig: &humayHTTPServer.HTTPConfig{
			Host: "localhost",
			Port: 8080,
			ReadTimeout: 5,
			WriteTimeout: 10,
			IdleTimeout: 20,
		},
	}

	app, err := serverApp.NewApp(config)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.Run(ctx)
}

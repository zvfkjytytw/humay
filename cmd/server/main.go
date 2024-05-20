package main

import (
	"context"
	"flag"

	serverApp "github.com/zvfkjytytw/humay/internal/server/app"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./build/server.yaml", "Server config file")
	flag.Parse()

	app, err := serverApp.NewAppFromFile(configFile)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.Run(ctx)
}

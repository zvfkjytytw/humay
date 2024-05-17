package main

import (
	"context"
	"flag"

	humayApp "github.com/zvfkjytytw/humay/internal/server/app"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./conf/server.yaml", "Server config file")
	flag.Parse()

	config, err := humayApp.GetConfig(configFile)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := humayApp.NewApp(config)
	if err != nil {
		panic(err)
	}

	err = app.Run(ctx)
	if err != nil {
		panic(err)
	}
}

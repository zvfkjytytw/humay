package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"strings"

	serverApp "github.com/zvfkjytytw/humay/internal/server/app"
	humayHTTPServer "github.com/zvfkjytytw/humay/internal/server/http"
)

const envAddress = "ADDRESS"

func main() {
	var configFile string
	var address string
	flag.StringVar(&configFile, "c", "./build/server.yaml", "Server config file")
	flag.StringVar(&address, "a", "localhost:8080", "Server address")
	flag.Parse()

	value, ok := os.LookupEnv(envAddress)
	if ok {
		address = value
	}
	host, port := splitAddress(address)

	config := &serverApp.ServerConfig{
		HTTPConfig: &humayHTTPServer.HTTPConfig{
			Host:         host,
			Port:         port,
			ReadTimeout:  5,
			WriteTimeout: 10,
			IdleTimeout:  20,
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

func splitAddress(address string) (host string, port int32) {
	values := strings.Split(address, ":")
	if len(values) == 1 {
		host = values[0]
		port = 8080
		return
	}

	host = values[0]
	p, err := strconv.Atoi(values[1])
	if err != nil {
		port = 8080
	}
	port = int32(p)

	return
}

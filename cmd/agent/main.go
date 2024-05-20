package main

import (
	"context"
	"flag"
	"strconv"
	"strings"

	agentApp "github.com/zvfkjytytw/humay/internal/agent/app"
)

func main() {
	var configFile, address string
	var pollInterval, reportInterval int
	flag.StringVar(&configFile, "c", "./build/agent.yaml", "Agent config file")
	flag.StringVar(&address, "a", "localhost:8080", "Server address")
	flag.IntVar(&pollInterval, "p", 2, "Interval for polling metrics")
	flag.IntVar(&reportInterval, "r", 10, "Interval for reporting metrics")
	flag.Parse()

	host, port := splitAddress(address)

	config := &agentApp.AgentConfig{
		ServerAddress:  host,
		ServerPort:     port,
		ServerType:     "http",
		PollInterval:   int32(pollInterval),
		ReportInterval: int32(reportInterval),
	}

	app, err := agentApp.NewApp(config)
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

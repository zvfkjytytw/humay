package main

import (
	"context"
	"flag"

	agentApp "github.com/zvfkjytytw/humay/internal/agent/app"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "c", "./build/agent.yaml", "Agent config file")
	flag.Parse()

	config := &agentApp.AgentConfig{
		ServerAddress:  "localhost",
		ServerPort:     8080,
		ServerType:     "http",
		PollInterval:   2,
		ReportInterval: 10,
	}

	app, err := agentApp.NewApp(config)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.Run(ctx)
}

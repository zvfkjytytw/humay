package main

import (
	"context"
	"flag"
	"os"
	"strconv"
	"strings"

	agentApp "github.com/zvfkjytytw/humay/internal/agent/app"
)

const (
	envAddress        = "ADDRESS"
	envReportInterval = "REPORT_INTERVAL"
	envPollInterval   = "POLL_INTERVAL"
	envRateLimit      = "RATE_LIMIT"
	envKey            = "KEY"
)

func main() {
	var (
		// Agent config file
		configFile string
		// Server address and port
		address string
		// Interval for polling metrics
		pollInterval int
		// Interval for send metrics to server
		reportInterval int
		// Limit for open connections to server
		rateLimit int
		// hash key for signature
		hashKey string
	)
	flag.StringVar(&configFile, "c", "./build/agent.yaml", "Agent config file")
	flag.StringVar(&address, "a", "localhost:8080", "Server address")
	flag.IntVar(&pollInterval, "p", 2, "Interval for polling metrics")
	flag.IntVar(&reportInterval, "r", 10, "Interval for reporting metrics")
	flag.IntVar(&rateLimit, "l", 5, "Rate limit")
	flag.StringVar(&hashKey, "k", "", "Key for generate hash")
	flag.Parse()

	value, ok := os.LookupEnv(envAddress)
	if ok {
		address = value
	}
	host, port := splitAddress(address)

	value, ok = os.LookupEnv(envPollInterval)
	if ok {
		interval, err := strconv.Atoi(value)
		if err == nil {
			pollInterval = interval
		}
	}

	value, ok = os.LookupEnv(envReportInterval)
	if ok {
		interval, err := strconv.Atoi(value)
		if err == nil {
			reportInterval = interval
		}
	}

	value, ok = os.LookupEnv(envRateLimit)
	if ok {
		limit, err := strconv.Atoi(value)
		if err == nil {
			rateLimit = limit
		}
	}

	value, ok = os.LookupEnv(envKey)
	if ok {
		hashKey = value
	}

	config := &agentApp.AgentConfig{
		ServerAddress:  host,
		ServerPort:     port,
		ServerType:     "http",
		PollInterval:   int32(pollInterval),
		ReportInterval: int32(reportInterval),
		RateLimit:      int32(rateLimit),
		HashKey:        hashKey,
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

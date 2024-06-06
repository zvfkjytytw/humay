package humayagent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	agentHTTP "github.com/zvfkjytytw/humay/internal/agent/http"
	metrics "github.com/zvfkjytytw/humay/internal/agent/metrics"
	common "github.com/zvfkjytytw/humay/internal/common"
)

type serverClient interface {
	UpdateGauge(metricName string, metricValue float64) error
	UpdateCounter(metricName string, metricValue int64) error
	Stop()
}

type AgentConfig struct {
	ServerAddress  string `yaml:"server_address"`
	ServerPort     int32  `yaml:"server_port"`
	ServerType     string `yaml:"server_type"`
	PollInterval   int32  `yaml:"poll_interval"`
	ReportInterval int32  `yaml:"report_interval"`
}

type AgentApp struct {
	pollInterval   int32
	reportInterval int32
	client         serverClient
	poller         *metrics.Poller
	logger         *zap.Logger
}

func NewApp(config *AgentConfig) (*AgentApp, error) {
	// Init logger
	logger, err := common.InitLogger()
	if err != nil {
		return nil, err
	}

	// Init metrics poller
	poller, err := metrics.NewPoller()
	if err != nil {
		return nil, err
	}
	poller.FlushPollCount()

	// Init server client
	var client serverClient
	if config.ServerType == "http" {
		client, err = agentHTTP.NewClient(fmt.Sprintf("%s:%d", config.ServerAddress, config.ServerPort), logger)
		if err != nil {
			return nil, err
		}
	}

	return &AgentApp{
		pollInterval:   config.PollInterval,
		reportInterval: config.ReportInterval,
		client:         client,
		poller:         poller,
		logger:         logger,
	}, nil
}

func NewAppFromFile(configFile string) (*AgentApp, error) {
	config := &AgentConfig{}
	configData, err := common.ReadConfigFile(configFile)
	if err != nil {
		return nil, err //nolint //wraped higher
	}

	err = yaml.Unmarshal(configData, config)
	if err != nil {
		return nil, err //nolint //wraped higher
	}

	return NewApp(config)
}

func (a *AgentApp) Run(ctx context.Context) {
	pollTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer reportTicker.Stop()

	sigChanel := make(chan os.Signal, 1)
	signal.Notify(sigChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	defer signal.Stop(sigChanel)
	defer a.client.Stop()

	for {
		select {
		case <-sigChanel:
			return
		case <-pollTicker.C:
			a.poller.Update()
		case <-reportTicker.C:
			func() {
				for metricName, metricValue := range a.poller.Metrics.Gauge {
					a.client.UpdateGauge(metricName, metricValue)
				}
				for metricName, metricValue := range a.poller.Metrics.Counter {
					a.client.UpdateCounter(metricName, metricValue)
				}
				a.poller.FlushPollCount()
			}()
		}
	}
}

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
	httpModels "github.com/zvfkjytytw/humay/internal/common/http/models"
)

const batchSize = 5

type serverClient interface {
	UpdateGauge(metricName string, metricValue float64) error
	UpdateCounter(metricName string, metricValue int64) error
	UpdateJSONGauge(metricName string, metricValue float64) error
	UpdateJSONCounter(metricName string, metricValue int64) error
	UpdateJSONMetrics([]*httpModels.Metric) error
	Stop()
}

type AgentConfig struct {
	ServerAddress  string `yaml:"server_address"`
	ServerPort     int32  `yaml:"server_port"`
	ServerType     string `yaml:"server_type"`
	PollInterval   int32  `yaml:"poll_interval"`
	ReportInterval int32  `yaml:"report_interval"`
	HashKey        string `yaml:"hash_key"`
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
		client, err = agentHTTP.NewClient(
			fmt.Sprintf("%s:%d", config.ServerAddress, config.ServerPort),
			logger,
			config.HashKey,
		)
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
			a.batchReport()
			// a.simpleReport()
		}
	}
}

func (a *AgentApp) simpleReport() {
	defer a.poller.FlushPollCount()
	for metricName, metricValue := range a.poller.Metrics.Gauge {
		if err := a.client.UpdateJSONGauge(metricName, metricValue); err != nil {
			a.logger.Sugar().Errorf("failed update gauge metric %s: %v", metricName, err)
		}
	}
	for metricName, metricValue := range a.poller.Metrics.Counter {
		if err := a.client.UpdateJSONCounter(metricName, metricValue); err != nil {
			a.logger.Sugar().Errorf("failed update gauge metric %s: %v", metricName, err)
		}
	}
}

func (a *AgentApp) batchReport() {
	defer a.poller.FlushPollCount()
	var metrics []*httpModels.Metric
	for metricName, metricValue := range a.poller.Metrics.Gauge {
		metrics = append(
			metrics,
			&httpModels.Metric{
				ID:    metricName,
				MType: "gauge",
				Value: &metricValue,
			},
		)
	}
	for metricName, metricValue := range a.poller.Metrics.Counter {
		metrics = append(
			metrics,
			&httpModels.Metric{
				ID:    metricName,
				MType: "counter",
				Delta: &metricValue,
			},
		)
	}

	limitIndex := batchSize * (len(metrics) / batchSize)
	for i := 0; i < limitIndex; i += batchSize {
		if err := a.client.UpdateJSONMetrics(metrics[i : i+batchSize]); err != nil {
			a.logger.Sugar().Errorf("failed update metrics: %v", err)
		}
	}

	if err := a.client.UpdateJSONMetrics(metrics[limitIndex:]); err != nil {
		a.logger.Sugar().Errorf("failed update metrics: %v", err)
	}
}

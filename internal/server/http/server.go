package humayhttpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Storage interface {
	GetGaugeMetric(name string) (float64, error)
	PutGaugeMetric(name string, value float64) error
	PutGaugeMetrics(map[string]float64) error
	GetCounterMetric(name string) (int64, error)
	PutCounterMetric(name string, value int64) error
	PutCounterMetrics(map[string]int64) error
	GetAllMetrics() map[string]map[string]string
	CheckDBConnect() error
	GetType() string
	Close() error
}

type HTTPConfig struct {
	Host         string `yaml:"host"`
	Port         int32  `yaml:"port"`
	ReadTimeout  int32  `yaml:"read_timeout"`
	WriteTimeout int32  `yaml:"write_timeout"`
	IdleTimeout  int32  `yaml:"idle_timeout"`
	HashKey      string `yaml:"hash_key"`
}

type HTTPServer struct {
	server  *http.Server
	logger  *zap.Logger
	storage Storage
	hashKey string
}

func NewHTTPServer(
	config *HTTPConfig,
	comlog *zap.Logger,
	storage Storage,
) *HTTPServer {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
	}

	logger, err := initLogger(comlog)
	if err != nil {
		comlog.Sugar().Errorf("failed init http logger: %v", err)
		logger = comlog
	}

	return &HTTPServer{
		server:  server,
		logger:  logger,
		storage: storage,
		hashKey: config.HashKey,
	}
}

func (h *HTTPServer) Start(ctx context.Context) error {
	router := h.newRouter()
	h.server.Handler = router

	err := h.server.ListenAndServe()
	if err != nil {
		h.logger.Sugar().Errorf("failed start http server: %w", err)
		return err
	}

	fmt.Println("http server started")

	return nil
}

func (h *HTTPServer) Stop(ctx context.Context) error {
	defer h.logger.Sync()
	err := h.server.Shutdown(ctx)
	if err != nil {
		h.logger.Sugar().Errorf("failed stop http server: %w", err)
		return err
	}

	err = h.storage.Close()
	if err != nil {
		h.logger.Sugar().Errorf("failed close storage: %w", err)
		return err
	}

	return nil
}

func initLogger(comlog *zap.Logger) (*zap.Logger, error) {
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(comlog.Level()),
		Development:      true,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout", "httpacc.log"},
		ErrorOutputPaths: []string{"stderr", "httperr.log"},
	}
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

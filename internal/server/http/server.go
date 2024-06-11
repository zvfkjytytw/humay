package humayhttpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type MemStorage interface {
	GetGaugeMetric(name string) (float64, error)
	PutGaugeMetric(name string, value float64) error
	GetCounterMetric(name string) (int64, error)
	PutCounterMetric(name string, value int64) error
	GetAllMetrics() map[string]map[string]string
}

type HTTPConfig struct {
	Host         string `yaml:"host"`
	Port         int32  `yaml:"port"`
	ReadTimeout  int32  `yaml:"read_timeout"`
	WriteTimeout int32  `yaml:"write_timeout"`
	IdleTimeout  int32  `yaml:"idle_timeout"`
}

type HTTPServer struct {
	server  *http.Server
	logger  *zap.Logger
	storage MemStorage
}

func NewHTTPServer(
	config *HTTPConfig,
	logger *zap.Logger,
	storage MemStorage,
) *HTTPServer {
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
	}

	return &HTTPServer{
		server:  server,
		logger:  logger,
		storage: storage,
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
	err := h.server.Shutdown(ctx)
	if err != nil {
		h.logger.Sugar().Errorf("failed stop http server: %w", err)
		return err
	}

	return nil
}

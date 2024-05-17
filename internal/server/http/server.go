package humayhttpserver

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	humayStorage "github.com/zvfkjytytw/humay/internal/server/storage"
)

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port int32  `yaml:"port"`
}

type HTTPServer struct {
	config  *HTTPConfig
	logger  *zap.Logger
	storage *humayStorage.MemStorage
}

func NewHTTPServer(
	config *HTTPConfig,
	logger *zap.Logger,
	storage *humayStorage.MemStorage,
) *HTTPServer {
	return &HTTPServer{
		config:  config,
		logger:  logger,
		storage: storage,
	}
}

func (h *HTTPServer) Start(ctx context.Context) error {
	router := h.newRouter()

	http.ListenAndServe(fmt.Sprintf("%s:%d", h.config.Host, h.config.Port), router)

	return nil
}

func (h *HTTPServer) Stop() error {
	return nil
}

package humayapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	humayHTTPServer "github.com/zvfkjytytw/humay/internal/server/http"
	humayStorage "github.com/zvfkjytytw/humay/internal/server/storage"
)

type Service interface {
	Start(ctx context.Context) error
	Stop() error
}

type AppConfig struct {
	httpConfig *humayHTTPServer.HTTPConfig `yaml:"http_config"`
}

type App struct {
	Logger   *zap.Logger
	Services []Service
}

func NewApp(config *AppConfig) (*App, error) {
	// Init logger
	logger, err := initLogger()
	if err != nil {
		return nil, err
	}

	// Init storage
	storage := humayStorage.NewStorage()

	// Init HTTP server
	httpServer := humayHTTPServer.NewHTTPServer(config.httpConfig, logger, storage)

	return &App{
		Logger: logger,
		Services: []Service{
			httpServer,
		},
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	for _, service := range a.Services {
		err := service.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed start service %w", err)
		}
		wg.Add(1)
	}

	stopSignal := <-signalChanel
	a.Logger.Error(fmt.Sprintf("Stop by %v", stopSignal))

	for _, service := range a.Services {
		err := service.Stop()
		if err != nil {
			a.Logger.Error(fmt.Sprintf("Stop service fail: %w", err))
		}
		wg.Done()
	}
	
	wg.Wait()

	err := a.Logger.Sync()
	if err != nil {
		return fmt.Errorf("failed start service %w", err)
	}

	return nil
}

func initLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

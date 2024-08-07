package humayserver

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	common "github.com/zvfkjytytw/humay/internal/common"
	humayHTTPServer "github.com/zvfkjytytw/humay/internal/server/http"
	humayStorage "github.com/zvfkjytytw/humay/internal/server/storage"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ServerConfig struct {
	HTTPConfig  *humayHTTPServer.HTTPConfig `yaml:"http_config" json:"http_config"`
	SaverConfig *SaverConfig                `yaml:"saver_config" json:"saver_config"`
	DatabaseDSN string                      `yaml:"database_dsn" json:"database_dsn"`
}

type ServerApp struct {
	logger   *zap.Logger
	services []Service
}

func NewApp(config *ServerConfig) (*ServerApp, error) {
	// Init logger
	logger, err := common.InitLogger()
	if err != nil {
		return nil, err
	}

	// Init App
	app := &ServerApp{
		logger: logger,
	}

	// Init storage
	var storage humayHTTPServer.Storage

	storage, err = humayStorage.NewPGStorage(config.DatabaseDSN)
	if err != nil {
		logger.Sugar().Errorf("failed init postgres storage% %v", err)
		memStorage := humayStorage.NewStorage(config.SaverConfig.StorageFile, config.DatabaseDSN)
		if config.SaverConfig.Restore {
			err := memStorage.Restore(config.SaverConfig.StorageFile)
			if err != nil {
				logger.Sugar().Errorf("failed restore storage from file %s: %v", config.SaverConfig.StorageFile, err)
			}
		}
		// Init Saver
		if config.SaverConfig.Interval == 0 {
			memStorage.SetAutoSave()
		} else {
			saver := newSaver(memStorage, config.SaverConfig.Interval, logger)
			app.services = append(app.services, saver)
		}

		storage = memStorage
	}

	// Init HTTP server
	httpServer := humayHTTPServer.NewHTTPServer(config.HTTPConfig, logger, storage)
	app.services = append(app.services, httpServer)

	return app, nil
}

func NewAppFromFile(configFile string) (*ServerApp, error) {
	config := &ServerConfig{}
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

func (a *ServerApp) Run(ctx context.Context) {
	defer a.logger.Sync()
	sigChanel := make(chan os.Signal, 1)
	signal.Notify(sigChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	for _, service := range a.services {
		go func(service Service) {
			err := service.Start(ctx)
			if err != nil {
				a.logger.Sugar().Errorf("service not started: %w", err)
				a.StopAll(ctx)
				return
			}
		}(service)
	}

	stopSignal := <-sigChanel
	a.logger.Sugar().Debugf("Stop by %v", stopSignal)
	a.StopAll(ctx)
}

func (a *ServerApp) StopAll(ctx context.Context) {
	for _, service := range a.services {
		err := service.Stop(ctx)
		if err != nil {
			a.logger.Error("stop failed")
		}
	}
}

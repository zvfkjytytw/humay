package humayserver

import (
	"context"
	"time"

	"go.uber.org/zap"

	humayStorage "github.com/zvfkjytytw/humay/internal/server/storage"
)

type SaverConfig struct {
	Interval    int32  `yaml:"interval,omitempty" json:"interval,omitempty"`
	StorageFile string `yaml:"storage_file" json:"storage_file"`
	Restore     bool   `yaml:"restore,omitempty" json:"restore,omitempty"`
}

type saver struct {
	storage  *humayStorage.MemStorage
	interval int32
	done     chan struct{}
	logger   *zap.Logger
}

func newSaver(
	storage *humayStorage.MemStorage,
	interval int32,
	logger *zap.Logger,
) *saver {
	return &saver{
		storage:  storage,
		interval: interval,
		done:     make(chan struct{}),
		logger:   logger,
	}
}

func (s *saver) Start(ctx context.Context) error {
	saveTicker := time.NewTicker(time.Duration(s.interval) * time.Second)
	defer saveTicker.Stop()

	for {
		select {
		case <-saveTicker.C:
			err := s.storage.Save()
			if err != nil {
				s.logger.Sugar().Errorf("failed save data: %v", err)
			} else {
				s.logger.Error("data saved")
			}
		case <-s.done:
			return nil
		}
	}
}

func (s *saver) Stop(ctx context.Context) error {
	close(s.done)
	return nil
}

package humaystorage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/sethvargo/go-retry"
)

const (
	saveTimeout    = 10 * time.Second
	expectIncrease = 2 * time.Second
	startExpect    = 1 * time.Second
	maxRetries     = 4
)

func (s *MemStorage) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
	defer cancel()
	backoff := retry.WithMaxRetries(
		maxRetries,
		retry.WithCappedDuration(
			expectIncrease,
			retry.NewFibonacci(startExpect),
		),
	)

	if err := retry.Do(
		ctx,
		backoff,
		func(ctx context.Context) error {
			s.mx.RLock()
			buf, err := json.Marshal(s)
			if err != nil {
				return err
			}
			s.mx.RUnlock()

			file, err := os.OpenFile(s.storageFile, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return err
			}
			defer file.Close()

			n, err := file.Write(buf)
			if err != nil {
				return err
			}

			if n == 0 {
				return errors.New("nothing is saved")
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

func (s *MemStorage) Restore(filePath string) error {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, s)
}

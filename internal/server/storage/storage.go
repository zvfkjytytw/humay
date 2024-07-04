package humaystorage

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemStorage struct {
	mx             sync.RWMutex
	autosave       bool
	storageType    string
	storageFile    string
	GaugeMetrics   map[string]float64 `json:"gauge_metrics,omitempty"`
	CounterMetrics map[string]int64   `json:"counter_metrics,omitempty"`
	dsn            string
}

func NewStorage(storageFile string, dsn string) *MemStorage {
	return &MemStorage{
		autosave:       false,
		storageType:    "struct",
		storageFile:    storageFile,
		GaugeMetrics:   make(map[string]float64),
		CounterMetrics: make(map[string]int64),
		dsn:            dsn,
	}
}

func (s *MemStorage) GetType() string {
	return s.storageType
}

func (s *MemStorage) SetAutoSave() {
	s.autosave = true
}

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) GetGaugeMetric(name string) (value float64, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.GaugeMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s is not found", name)
	}

	return
}

func (s *MemStorage) PutGaugeMetric(name string, value float64) (err error) {
	defer func() {
		if s.autosave {
			s.Save()
		}
	}()
	s.mx.Lock()
	defer s.mx.Unlock()
	s.GaugeMetrics[name] = value

	return
}

func (s *MemStorage) GetCounterMetric(name string) (value int64, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.CounterMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s not found", name)
	}

	return
}

func (s *MemStorage) PutCounterMetric(name string, value int64) (err error) {
	defer func() {
		if s.autosave {
			s.Save()
		}
	}()
	s.mx.Lock()
	defer s.mx.Unlock()
	s.CounterMetrics[name] = s.CounterMetrics[name] + value

	return
}

func (s *MemStorage) GetAllMetrics() map[string]map[string]string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metrics := make(map[string]map[string]string)
	metrics["gauges"] = make(map[string]string)

	for name, value := range s.GaugeMetrics {
		metrics["gauges"][name] = strconv.FormatFloat(value, 'f', -1, 64)
	}

	metrics["counters"] = make(map[string]string)
	for name, value := range s.CounterMetrics {
		metrics["counters"][name] = strconv.FormatInt(value, 10)
	}

	return metrics
}

func (s *MemStorage) CheckDBConnect() error {
	ctx := context.Background()
	config, err := pgx.ParseConfig(s.dsn)
	if err != nil {
		return err
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgxpool.New(ctx, config.ConnString())
	fmt.Println(s.dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = conn.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *MemStorage) PutGaugeMetrics(metrics map[string]float64) (err error) {
	for name, value := range metrics {
		err = s.PutGaugeMetric(name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MemStorage) PutCounterMetrics(metrics map[string]int64) (err error) {
	for name, value := range metrics {
		err = s.PutCounterMetric(name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

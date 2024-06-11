package humaystorage

import (
	"fmt"
	"strconv"
	"sync"
)

type MemStorage struct {
	storageType    string
	mx             sync.RWMutex
	gaugeMetrics   map[string]float64
	counterMetrics map[string]int64
}

func NewStorage() *MemStorage {
	return &MemStorage{
		storageType:    "struct",
		gaugeMetrics:   make(map[string]float64),
		counterMetrics: make(map[string]int64),
	}
}

func (s *MemStorage) GetType() string {
	return s.storageType
}

func (s *MemStorage) GetGaugeMetric(name string) (value float64, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.gaugeMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s is not found", name)
	}

	return
}

func (s *MemStorage) PutGaugeMetric(name string, value float64) (err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.gaugeMetrics[name] = value

	return
}

func (s *MemStorage) GetCounterMetric(name string) (value int64, err error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	value, ok := s.counterMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s not found", name)
	}

	return
}

func (s *MemStorage) PutCounterMetric(name string, value int64) (err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.counterMetrics[name] = s.counterMetrics[name] + value

	return
}

func (s *MemStorage) GetAllMetrics() map[string]map[string]string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metrics := make(map[string]map[string]string)
	metrics["gauges"] = make(map[string]string)

	for name, value := range s.gaugeMetrics {
		metrics["gauges"][name] = strconv.FormatFloat(value, 'f', -1, 64)
	}

	metrics["counters"] = make(map[string]string)
	for name, value := range s.counterMetrics {
		metrics["counters"][name] = strconv.FormatInt(value, 10)
	}

	return metrics
}

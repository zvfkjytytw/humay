package humaystorage

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

type MemStorage struct {
	storageType    string
	mx             sync.RWMutex
	gaugeMetrics   map[string]*gaugeMetric
	counterMetrics map[string]*counterMetric
}

type gaugeMetric struct {
	mx     sync.RWMutex
	value  float64
	vector map[int64]float64
}

type counterMetric struct {
	mx     sync.RWMutex
	value  int64
	vector map[int64]int64
}

func NewStorage() *MemStorage {
	return &MemStorage{
		storageType:    "struct",
		gaugeMetrics:   make(map[string]*gaugeMetric),
		counterMetrics: make(map[string]*counterMetric),
	}
}

func (s *MemStorage) GetType() string {
	return s.storageType
}

func (s *MemStorage) GetGaugeMetric(name string) (value float64, err error) {
	metric, ok := s.gaugeMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s is not found", name)
		return
	}
	metric.mx.RLock()
	defer metric.mx.RUnlock()
	value = metric.value

	return
}

func (s *MemStorage) GetGaugeMetrics(name string) (value map[int64]float64, err error) {
	metric, ok := s.gaugeMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s is not found", name)
		return
	}
	metric.mx.RLock()
	defer metric.mx.RUnlock()
	value = metric.vector

	return
}

func (s *MemStorage) PutGaugeMetric(name string, value float64) (err error) {
	metric, ok := s.gaugeMetrics[name]
	if ok {
		metric.mx.Lock()
		defer metric.mx.Unlock()
		metric.value = value
		metric.vector[time.Now().Unix()] = value
	} else {
		s.mx.Lock()
		defer s.mx.Unlock()
		s.gaugeMetrics[name] = &gaugeMetric{
			value: value,
			vector: map[int64]float64{
				time.Now().Unix(): value,
			},
		}
	}

	return
}

func (s *MemStorage) GetCounterMetric(name string) (value int64, err error) {
	metric, ok := s.counterMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s not found", name)
		return
	}
	metric.mx.RLock()
	defer metric.mx.RUnlock()
	value = metric.value

	return
}

func (s *MemStorage) GetCounterMetrics(name string) (value map[int64]int64, err error) {
	metric, ok := s.counterMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s not found", name)
		return
	}
	metric.mx.RLock()
	defer metric.mx.RUnlock()
	value = metric.vector

	return
}

func (s *MemStorage) PutCounterMetric(name string, value int64) (err error) {
	metric, ok := s.counterMetrics[name]
	if ok {
		metric.mx.Lock()
		defer metric.mx.Unlock()
		metric.value += value
		metric.vector[time.Now().Unix()] = value
	} else {
		s.mx.Lock()
		defer s.mx.Unlock()
		s.counterMetrics[name] = &counterMetric{
			value: value,
			vector: map[int64]int64{
				time.Now().Unix(): value,
			},
		}
	}

	return
}

func (s *MemStorage) GetAllMetrics() map[string]map[string]string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metrics := make(map[string]map[string]string)
	metrics["gauges"] = make(map[string]string)

	for name, gauge := range s.gaugeMetrics {
		metrics["gauges"][name] = strconv.FormatFloat(gauge.value, 'f', -1, 64)
	}

	metrics["counters"] = make(map[string]string)
	for name, counter := range s.counterMetrics {
		metrics["counters"][name] = strconv.FormatInt(counter.value, 10)
	}

	return metrics
}

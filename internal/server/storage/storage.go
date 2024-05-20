package humaystorage

import (
	"fmt"
	"time"
)

type MemStorage struct {
	storageType    string
	gaugeMetrics   map[string]*gaugeMetric
	counterMetrics map[string]*counterMetric
}

type gaugeMetric struct {
	value  float64
	vector map[int64]float64
}

type counterMetric struct {
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
	value = metric.value

	return
}

func (s *MemStorage) GetGaugeMetrics(name string) (value map[int64]float64, err error) {
	metric, ok := s.gaugeMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s is not found", name)
		return
	}
	value = metric.vector

	return
}

func (s *MemStorage) PutGaugeMetric(name string, value float64) (err error) {
	metric, ok := s.gaugeMetrics[name]
	if ok {
		metric.value = value
		metric.vector[time.Now().Unix()] = value
	} else {
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
	value = metric.value

	return
}

func (s *MemStorage) GetCounterMetrics(name string) (value map[int64]int64, err error) {
	metric, ok := s.counterMetrics[name]
	if !ok {
		err = fmt.Errorf("metric %s not found", name)
		return
	}
	value = metric.vector

	return
}

func (s *MemStorage) PutCounterMetric(name string, value int64) (err error) {
	metric, ok := s.counterMetrics[name]
	if ok {
		metric.value += value
		metric.vector[time.Now().Unix()] = value
	} else {
		s.counterMetrics[name] = &counterMetric{
			value: value,
			vector: map[int64]int64{
				time.Now().Unix(): value,
			},
		}
	}

	return
}

package humaymetricspoller

import (
	"math/rand"
	"runtime"
)

type Metrics struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

type Poller struct {
	Metrics  *Metrics
	memStats *runtime.MemStats
}

func NewPoller() (*Poller, error) {
	return &Poller{
		Metrics: &Metrics{
			Gauge:   make(map[string]float64),
			Counter: make(map[string]int64),
		},
		memStats: &runtime.MemStats{},
	}, nil
}

func (p *Poller) Update() {
	// update memory statistics
	p.updateMemMetrics()

	// update gauge metrics
	p.Metrics.Gauge["Alloc"] = float64(p.memStats.Alloc)
	p.Metrics.Gauge["BuckHashSys"] = float64(p.memStats.BuckHashSys)
	p.Metrics.Gauge["Frees"] = float64(p.memStats.Frees)
	p.Metrics.Gauge["GCCPUFraction"] = float64(p.memStats.GCCPUFraction)
	p.Metrics.Gauge["GCSys"] = float64(p.memStats.GCSys)
	p.Metrics.Gauge["HeapAlloc"] = float64(p.memStats.HeapAlloc)
	p.Metrics.Gauge["HeapIdle"] = float64(p.memStats.HeapIdle)
	p.Metrics.Gauge["HeapInuse"] = float64(p.memStats.HeapInuse)
	p.Metrics.Gauge["HeapObjects"] = float64(p.memStats.HeapObjects)
	p.Metrics.Gauge["HeapReleased"] = float64(p.memStats.HeapReleased)
	p.Metrics.Gauge["HeapSys"] = float64(p.memStats.HeapSys)
	p.Metrics.Gauge["LastGC"] = float64(p.memStats.LastGC)
	p.Metrics.Gauge["Lookups"] = float64(p.memStats.Lookups)
	p.Metrics.Gauge["MCacheInuse"] = float64(p.memStats.MCacheInuse)
	p.Metrics.Gauge["MCacheSys"] = float64(p.memStats.MCacheSys)
	p.Metrics.Gauge["MSpanInuse"] = float64(p.memStats.MSpanInuse)
	p.Metrics.Gauge["MSpanSys"] = float64(p.memStats.MSpanSys)
	p.Metrics.Gauge["Mallocs"] = float64(p.memStats.Mallocs)
	p.Metrics.Gauge["NextGC"] = float64(p.memStats.NextGC)
	p.Metrics.Gauge["NumForcedGC"] = float64(p.memStats.NumForcedGC)
	p.Metrics.Gauge["NumGC"] = float64(p.memStats.NumGC)
	p.Metrics.Gauge["OtherSys"] = float64(p.memStats.OtherSys)
	p.Metrics.Gauge["PauseTotalNs"] = float64(p.memStats.PauseTotalNs)
	p.Metrics.Gauge["StackInuse"] = float64(p.memStats.StackInuse)
	p.Metrics.Gauge["StackSys"] = float64(p.memStats.StackSys)
	p.Metrics.Gauge["Sys"] = float64(p.memStats.Sys)
	p.Metrics.Gauge["TotalAlloc "] = float64(p.memStats.TotalAlloc)
	p.Metrics.Gauge["RandomValue"] = rand.Float64()

	// update counter metrics
	// p.Metrics.Counter["PollCount"] = int64(len(p.Metrics.Gauge))
	p.Metrics.Counter["PollCount"] += 1
}

func (p *Poller) updateMemMetrics() {
	runtime.ReadMemStats(p.memStats)
}

func (p *Poller) FlushPollCount() {
	p.Metrics.Counter["PollCount"] = 0
}

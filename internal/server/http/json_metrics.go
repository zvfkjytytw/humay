package humayhttpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	httpModels "github.com/zvfkjytytw/humay/internal/common/http/models"
)

// return metric structure with the actual value from the storage.
func (h *HTTPServer) getJSONValue(w http.ResponseWriter, r *http.Request) {
	// parsing request body.
	contentType, ok := r.Header["Content-Type"]
	if !ok || contentType[0] != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong Content-Type. Expect application/json"))
		return
	}

	defer r.Body.Close()
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed read body"))
		return
	}

	requestMetric := &httpModels.Metric{}
	err = json.Unmarshal(requestBody, requestMetric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed unmarshal body"))
		return
	}

	metricType := requestMetric.MType
	metricName := requestMetric.ID
	if !checkMetricType(metricType) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("wrong metric type %s", metricType)))
		return
	}

	metric, err := h.getMetricStruct(metricType, metricName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("metric %s not found", metricName)))
		return
	}

	body, err := json.Marshal(metric)
	if err != nil {
		h.logger.Sugar().Errorf("failed marshal metric %s: %w", metricName, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed marshal metric"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// save metric with the name and value from the request Body to the storage.
func (h *HTTPServer) putJSONValue(w http.ResponseWriter, r *http.Request) {
	// parse request body.
	contentType, ok := r.Header["Content-Type"]
	if !ok || contentType[0] != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong Content-Type. Expect application/json"))
		return
	}

	defer r.Body.Close()
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed read body"))
		return
	}

	requestMetric := &httpModels.Metric{}
	err = json.Unmarshal(requestBody, requestMetric)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed unmarshal body"))
		return
	}

	metricType := requestMetric.MType
	metricName := requestMetric.ID
	if !checkMetricType(metricType) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("wrong metric type %s", metricType)))
		return
	}

	// save metric.
	switch metricType {
	case httpModels.GaugeMetric:
		err := h.storage.PutGaugeMetric(metricName, *requestMetric.Value)
		if err != nil {
			h.logger.Sugar().Errorf("failed save %s metric %s: %w", httpModels.GaugeMetric, metricName, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save metric"))
			return
		}

	case httpModels.CounterMetric:
		err := h.storage.PutCounterMetric(metricName, *requestMetric.Delta)
		if err != nil {
			h.logger.Sugar().Errorf("failed save %s metric %s: %w", httpModels.CounterMetric, metricName, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save metric"))
			return
		}
	}

	// return saved metric.
	metric, err := h.getMetricStruct(metricType, metricName) //nolint // this metric just saved
	if err != nil {
		h.logger.Sugar().Errorf("failed get metric %s: %w", metricName, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed marshal metric"))
		return
	}

	body, err := json.Marshal(metric)
	if err != nil {
		h.logger.Sugar().Errorf("failed marshal metric %s: %w", metricName, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed marshal metric"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Storage-Type", h.storage.GetType())
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// get metric structure with the actual value from the storage.
func (h *HTTPServer) getMetricStruct(mType, mName string) (*httpModels.Metric, error) {
	metric := &httpModels.Metric{
		ID:    mName,
		MType: mType,
	}

	switch mType {
	case httpModels.GaugeMetric:
		value, err := h.storage.GetGaugeMetric(mName)
		if err != nil {
			h.logger.Sugar().Errorf("failed get metric: %w", err)
			return nil, err
		}
		metric.Value = &value
	case httpModels.CounterMetric:
		value, err := h.storage.GetCounterMetric(mName)
		if err != nil {
			h.logger.Sugar().Errorf("failed get metric: %w", err)
			return nil, err
		}
		metric.Delta = &value
	}

	return metric, nil
}

// checking the type of metric for compliance with acceptable.
func checkMetricType(mType string) bool {
	for _, t := range httpModels.MetricTypes {
		if mType == t {
			return true
		}
	}

	return false
}

// save metrics with the name and value from the request Body to the storage.
func (h *HTTPServer) putJSONValues(w http.ResponseWriter, r *http.Request) {
	contentType, ok := r.Header["Content-Type"]
	if !ok || contentType[0] != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong Content-Type. Expect application/json"))
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed read body"))
		return
	}

	var metrics []*httpModels.Metric
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("failed unmarshal body"))
		return
	}

	gaugeMetrics := make(map[string]float64)
	counterMetrics := make(map[string]int64)

	for _, metric := range metrics {
		switch strings.ToLower(strings.TrimSpace(metric.MType)) {
		case "gauge":
			gaugeMetrics[strings.TrimSpace(metric.ID)] = *metric.Value
		case "counter":
			delta, ok := counterMetrics[strings.TrimSpace(metric.ID)]
			if ok {
				counterMetrics[strings.TrimSpace(metric.ID)] = delta + *metric.Delta
			} else {
				counterMetrics[strings.TrimSpace(metric.ID)] = *metric.Delta
			}
		}
	}

	if len(counterMetrics) > 0 {
		if err = h.storage.PutCounterMetrics(counterMetrics); err != nil {
			h.logger.Sugar().Errorf("failed save %s metrics: %w", httpModels.CounterMetric, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save counter metrics"))
			return
		}
	}

	if len(gaugeMetrics) > 0 {
		if err = h.storage.PutGaugeMetrics(gaugeMetrics); err != nil {
			h.logger.Sugar().Errorf("failed save %s metrics: %w", httpModels.GaugeMetric, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save gauge metrics"))
			return
		}
	}

	gauges := make([]string, 0, len(gaugeMetrics))
	for id := range gaugeMetrics {
		gauges = append(gauges, id)
	}

	counters := make([]string, 0, len(counterMetrics))
	for id := range counterMetrics {
		counters = append(counters, id)
	}

	metricsList, err := h.getMetricsList(gauges, counters)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("can't get saved metrics"))
		return
	}

	respBody, err := json.Marshal(metricsList)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("can't get saved metrics"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (h *HTTPServer) getMetricsList(gauges, counters []string) (metrics []*httpModels.Metric, err error) {
	metrics = make([]*httpModels.Metric, 0, len(gauges)+len(counters))

	for _, name := range gauges {
		value, err := h.storage.GetGaugeMetric(name)
		if err != nil {
			h.logger.Sugar().Errorf("failed get %s metric %s: %w", httpModels.GaugeMetric, name, err)
			return nil, err
		}
		metrics = append(
			metrics,
			&httpModels.Metric{
				ID:    name,
				MType: "gauge",
				Value: &value,
			},
		)
	}

	for _, name := range counters {
		value, err := h.storage.GetCounterMetric(name)
		if err != nil {
			h.logger.Sugar().Errorf("failed get %s metric %s: %w", httpModels.CounterMetric, name, err)
			return nil, err
		}
		metrics = append(
			metrics,
			&httpModels.Metric{
				ID:    name,
				MType: "counter",
				Delta: &value,
			},
		)
	}

	return metrics, nil
}

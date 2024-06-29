package humayhttpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	httpModels "github.com/zvfkjytytw/humay/internal/common/http/models"
)

// middleware for checking request condition and correctness of the body.
func jsonCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		metric := &httpModels.Metric{}
		err = json.Unmarshal(body, metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed unmarshal body"))
			return
		}

		if !checkMetricType(metric.MType) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("wrong metric type %s", metric.MType)))
			return
		}
		var mValue string
		if r.RequestURI == httpModels.UpdateHandler {
			switch metric.MType {
			case httpModels.GaugeMetric:
				if metric.Value == nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("not specified gauge value"))
					return
				}
				mValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			case httpModels.CounterMetric:
				if metric.Delta == nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("not specified counter delta"))
					return
				}
				mValue = strconv.FormatInt(*metric.Delta, 10)
			}
		}

		ctx := context.WithValue(r.Context(), contextMetricType, strings.TrimSpace(metric.MType))
		ctx = context.WithValue(ctx, contextMetricName, strings.TrimSpace(metric.ID))
		ctx = context.WithValue(ctx, contextMetricValue, mValue)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// return metric structure with the actual value from the storage.
func (h *HTTPServer) getJSONValue(w http.ResponseWriter, r *http.Request) {
	metricType := fmt.Sprintf("%v", r.Context().Value(contextMetricType))
	metricName := fmt.Sprintf("%v", r.Context().Value(contextMetricName))

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
	metricType := fmt.Sprintf("%v", r.Context().Value(contextMetricType))
	metricName := fmt.Sprintf("%v", r.Context().Value(contextMetricName))
	metricValue := fmt.Sprintf("%v", r.Context().Value(contextMetricValue))

	switch metricType {
	case httpModels.GaugeMetric:
		value, _ := strconv.ParseFloat(metricValue, 64) //nolint // wraped in middleware
		err := h.storage.PutGaugeMetric(metricName, value)
		if err != nil {
			h.logger.Sugar().Errorf("failed save %s metric %s: %w", httpModels.GaugeMetric, metricName, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save metric"))
			return
		}

	case httpModels.CounterMetric:
		value, _ := strconv.ParseInt(metricValue, 10, 64) //nolint // wraped in middleware
		err := h.storage.PutCounterMetric(metricName, value)
		if err != nil {
			h.logger.Sugar().Errorf("failed save %s metric %s: %w", httpModels.CounterMetric, metricName, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save metric"))
			return
		}
	}

	metric, _ := h.getMetricStruct(metricType, metricName) //nolint // this metric just saved
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
			counterMetrics[strings.TrimSpace(metric.ID)] = *metric.Delta
		}
	}

	if len(counterMetrics) > 0 {
		if err = h.storage.PutCounterMetrics(counterMetrics); err != nil {
			// h.logger.Sugar().Errorf("failed save %s metrics: %w", httpModels.CounterMetric, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed save counter metrics"))
			return
		}
	}

	if len(gaugeMetrics) > 0 {
		if err = h.storage.PutGaugeMetrics(gaugeMetrics); err != nil {
			// h.logger.Sugar().Errorf("failed save %s metrics: %w", httpModels.GaugeMetric, err)
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
			// h.logger.Sugar().Errorf("failed get %s metric %s: %w", httpModels.GaugeMetric, name, err)
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
			// h.logger.Sugar().Errorf("failed get %s metric %s: %w", httpModels.CounterMetric, name, err)
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

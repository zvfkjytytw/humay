package humayhttpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type contextKey int

const (
	contextMetricType contextKey = iota
	contextMetricName
	contextMetricValue
)

var metricsMap = map[string][]string{
	"gauge": {
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		"RandomValue",
		"testGauge",
	},
	"counter": {
		"PollCount",
		"testCounter",
	},
}

// checking URL path for correctness of the conditions for getting the metric value.
func valueCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if err := checkMetricName(metricType, metricName); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), contextMetricType, metricType)
		ctx = context.WithValue(ctx, contextMetricName, metricName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *HTTPServer) getValue(w http.ResponseWriter, r *http.Request) {
	metricType := fmt.Sprintf("%v", r.Context().Value(contextMetricType))
	metricName := fmt.Sprintf("%v", r.Context().Value(contextMetricName))
	var value string

	if metricType == "gauge" {
		v, err := h.storage.GetGaugeMetric(metricName)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		}
		value = strconv.FormatFloat(v, 'f', -1, 64)
	}

	if metricType == "counter" {
		v, err := h.storage.GetCounterMetric(metricName)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		}
		value = strconv.FormatInt(v, 10)
	}
	fmt.Printf("metric %v return value %v\n", metricName, value)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

// checking URL path for correctness of the conditions for saving the metric.
func updateCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if err := checkUpdateContext(metricType, metricValue); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}
		if err := checkMetricName(metricType, metricName); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), contextMetricType, metricType)
		ctx = context.WithValue(ctx, contextMetricName, metricName)
		ctx = context.WithValue(ctx, contextMetricValue, metricValue)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *HTTPServer) putValue(w http.ResponseWriter, r *http.Request) {
	metricType := fmt.Sprintf("%v", r.Context().Value(contextMetricType))
	metricName := fmt.Sprintf("%v", r.Context().Value(contextMetricName))
	metricValue := fmt.Sprintf("%v", r.Context().Value(contextMetricValue))

	if metricType == "gauge" {
		value, _ := strconv.ParseFloat(metricValue, 64) //nolint // wraped in checkUpdateContext
		h.storage.PutGaugeMetric(metricName, value)
	}
	if metricType == "counter" {
		value, _ := strconv.ParseInt(metricValue, 10, 64) //nolint // wraped in checkUpdateContext
		h.storage.PutCounterMetric(metricName, value)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("metric %s saved", metricName)))
}

// checking the URL for the correct metric type and value.
func checkUpdateContext(metricType, metricValue string) (err error) {
	if metricType == "gauge" {
		_, err = strconv.ParseFloat(metricValue, 64)
		if err == nil {
			return nil
		}

		return errors.New("wrong gauge value")
	}

	if metricType == "counter" {
		_, err = strconv.ParseInt(metricValue, 10, 64)
		if err == nil {
			return nil
		}

		return errors.New("wrong counter value")
	}

	return errors.New("unknown metric type")
}

// checking the URL for the correct metric name.
func checkMetricName(metricType, metricName string) error {
	metrics, ok := metricsMap[metricType]
	if !ok {
		return errors.New("unknown metric type")
	}

	for _, metric := range metrics {
		if metric == metricName {
			return nil
		}
	}

	return errors.New("unknown metric name")
}

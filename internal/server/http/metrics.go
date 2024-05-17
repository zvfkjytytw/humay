package humayhttpserver

import (
	"context"
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
	},
	"counter": {
		"PollCount",
	},
}

// checking URL path for correctness of the conditions for getting the metric value.
func metricCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if err := checkMetricName(metricType, metricName); err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), contextMetricType, metricType)
		ctx = context.WithValue(ctx, contextMetricName, metricName)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *HTTPServer) getValue(w http.ResponseWriter, r *http.Request) {
	metricType := r.Context().Value("type").(string)
	metricName := r.Context().Value("name").(string)
	var value string

	if metricType == "gauge" {
		v, err := h.storage.GetGaugeMetric(metricName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		value = strconv.FormatFloat(v, 'E', -1, 64)
	}

	if metricType == "counter" {
		v, err := h.storage.GetCounterMetric(metricName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		value = strconv.FormatInt(v, 10)
	}

	w.Write([]byte(value))
	w.WriteHeader(http.StatusOK)
}

// checking URL path for correctness of the conditions for saving the metric.
func updateCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		if err := checkUpdateContext(metricType, metricValue); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err := checkMetricName(metricType, metricName); err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), contextMetricType, metricType)
		ctx = context.WithValue(ctx, contextMetricName, metricName)
		ctx = context.WithValue(ctx, contextMetricValue, metricValue)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *HTTPServer) putValue(w http.ResponseWriter, r *http.Request) {
	metricType := r.Context().Value("type").(string)
	metricName := r.Context().Value("name").(string)
	metricValue := r.Context().Value("value").(string)

	if metricType == "gauge" {
		value, _ := strconv.ParseFloat(metricValue, 64)
		h.storage.PutGaugeMetric(metricName, value)
	}
	if metricType == "counter" {
		value, _ := strconv.ParseInt(metricValue, 10, 64)
		h.storage.PutCounterMetric(metricName, value)
	}

	w.Write([]byte(fmt.Sprintf("metric %s save", metricName)))
	w.WriteHeader(http.StatusOK)
}

// checking the URL for the correct metric type and value.
func checkUpdateContext(metricType, metricValue string) (err error) {
	if metricType == "gauge" {
		_, err = strconv.ParseFloat(metricValue, 64)
		if err == nil {
			return nil
		}

		return fmt.Errorf("wrong gauge value")
	}

	if metricType == "counter" {
		_, err = strconv.ParseInt(metricValue, 10, 64)
		if err == nil {
			return nil
		}

		return fmt.Errorf("wrong counter value")
	}

	return fmt.Errorf("unknown metric type")
}

// checking the URL for the correct metric name.
func checkMetricName(metricType, metricName string) error {
	metrics, ok := metricsMap[metricType]
	if !ok {
		return fmt.Errorf("unknown metric type")
	}

	for _, metric := range metrics {
		if metric == metricName {
			return nil
		}
	}

	return fmt.Errorf("unknown metric name")
}

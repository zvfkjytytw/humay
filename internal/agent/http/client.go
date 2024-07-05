package humayhttpagent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"

	httpModels "github.com/zvfkjytytw/humay/internal/common/http/models"
)

const (
	HTTPProtocol   = "http"
	HTTPSProtocol  = "https"
	requestTimeout = 10 * time.Second
	expectIncrease = 2 * time.Second
	startExpect    = 1 * time.Second
	maxRetries     = 4
)

type HTTPClient struct {
	address  string
	protocol string
	client   http.Client
	logger   *zap.Logger
}

func NewClient(address string, logger *zap.Logger) (*HTTPClient, error) {
	tr := &http.Transport{
		MaxIdleConns:    1,
		IdleConnTimeout: 60 * time.Second,
	}
	client := http.Client{Transport: tr}

	return &HTTPClient{
		address:  address,
		protocol: HTTPProtocol,
		client:   client,
		logger:   logger,
	}, nil
}

// update metric block for text/plain case.
func (h *HTTPClient) UpdateGauge(metricName string, metricValue float64) error {
	value := strconv.FormatFloat(metricValue, 'f', -1, 64)
	return h.updateMetric(httpModels.GaugeMetric, metricName, value) //nolint //wraped higher
}

func (h *HTTPClient) UpdateCounter(metricName string, metricValue int64) error {
	value := strconv.FormatInt(metricValue, 10)
	return h.updateMetric(httpModels.CounterMetric, metricName, value) //nolint //wraped higher
}

func (h *HTTPClient) updateMetric(metricType, metricName, metricValue string) error {
	var body string
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s://%s%s", h.protocol, h.address, strings.Join([]string{httpModels.UpdateHandler, metricType, metricName, metricValue}, "/")),
		strings.NewReader(body),
	)
	if err != nil {
		return err //nolint //wraped higher
	}

	req.Header.Add("Content-Type", "text/plain")
	resp, err := h.client.Do(req)
	if err != nil {
		return err //nolint //wraped higher
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("metric %s not saved", metricName)
	}

	return nil
}

// update metric block for application/json case.
func (h *HTTPClient) UpdateJSONGauge(metricName string, metricValue float64) error {
	metric := &httpModels.Metric{
		ID:    metricName,
		MType: httpModels.GaugeMetric,
		Value: &metricValue,
	}
	return h.updateJSONMetric(metric) //nolint //wraped higher
}

func (h *HTTPClient) UpdateJSONCounter(metricName string, metricValue int64) error {
	metric := &httpModels.Metric{
		ID:    metricName,
		MType: httpModels.CounterMetric,
		Delta: &metricValue,
	}
	return h.updateJSONMetric(metric) //nolint //wraped higher
}

func (h *HTTPClient) updateJSONMetric(metric *httpModels.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
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
			body, err := json.Marshal(metric)
			if err != nil {
				h.logger.Sugar().Errorf("failed marshal metric body: %v", metric)
				return err
			}

			buf := &bytes.Buffer{}
			gzWriter, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
			if _, err := gzWriter.Write(body); err != nil {
				h.logger.Sugar().Errorf("failed compress body: %v", err)
				return err
			}
			if err := gzWriter.Close(); err != nil {
				h.logger.Sugar().Errorf("failed close compressor: %v", err)
				return err
			}

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s://%s%s", h.protocol, h.address, httpModels.UpdateHandler),
				bytes.NewReader(buf.Bytes()),
			)
			if err != nil {
				return err //nolint //wraped higher
			}

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")
			resp, err := h.client.Do(req)
			if err != nil {
				return err //nolint //wraped higher
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("metric %s not saved", metric.ID)
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

func (h *HTTPClient) UpdateJSONMetrics(metrics []*httpModels.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
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
			body, err := json.Marshal(metrics)
			if err != nil {
				h.logger.Sugar().Errorf("failed marshal metric body: %v", metrics)
				return err
			}

			buf := &bytes.Buffer{}
			gzWriter, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
			if _, err := gzWriter.Write(body); err != nil {
				h.logger.Sugar().Errorf("failed compress body: %v", err)
				return err
			}
			if err := gzWriter.Close(); err != nil {
				h.logger.Sugar().Errorf("failed close compressor: %v", err)
				return err
			}

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s://%s%s", h.protocol, h.address, httpModels.UpdatesHandler),
				bytes.NewReader(buf.Bytes()),
			)
			if err != nil {
				return err //nolint //wraped higher
			}

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Content-Encoding", "gzip")
			resp, err := h.client.Do(req)
			if err != nil {
				return err //nolint //wraped higher
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					h.logger.Sugar().Errorf("failed read response body: %v", err)
				}
				return fmt.Errorf("metrics not saved: %s", string(bodyBytes))
			}
			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

func (h *HTTPClient) Stop() {
	h.client.CloseIdleConnections()
}

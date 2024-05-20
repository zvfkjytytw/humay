package humayhttpagent

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	HTTPProtocol  = "http:/"
	HTTPSProtocol = "https:/"
	updateHandler = "update"
	gaugeType     = "gauge"
	counterType   = "counter"
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

func (h *HTTPClient) UpdateGauge(metricName string, metricValue float64) error {
	value := strconv.FormatFloat(metricValue, 'f', -1, 64)
	return h.updateMetric(gaugeType, metricName, value) //nolint //wraped higher
}

func (h *HTTPClient) UpdateCounter(metricName string, metricValue int64) error {
	value := strconv.FormatInt(metricValue, 10)
	return h.updateMetric(counterType, metricName, value) //nolint //wraped higher
}

func (h *HTTPClient) updateMetric(metricType, metricName, metricValue string) error {
	body := ""

	req, err := http.NewRequest(
		http.MethodPost,
		strings.Join([]string{h.protocol, h.address, updateHandler, metricType, metricName, metricValue}, "/"),
		strings.NewReader(body),
	)
	if err != nil {
		return err //nolint //wraped higher
	}

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

func (h *HTTPClient) Stop() {
	h.client.CloseIdleConnections()
}

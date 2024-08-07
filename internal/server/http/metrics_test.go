package humayhttpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockStorage struct{}

func (m *mockStorage) GetGaugeMetric(name string) (value float64, err error) {
	if name == "fail" {
		err = errors.New("metric fail not found")
		return
	}

	return
}

func (m *mockStorage) PutGaugeMetric(name string, value float64) (err error) {
	if name == "fail" {
		err = errors.New("failed saved metric fail")
		return
	}

	return
}

func (m *mockStorage) GetCounterMetric(name string) (value int64, err error) {
	if name == "fail" {
		err = errors.New("metric fail not found")
		return
	}

	return
}

func (m *mockStorage) PutCounterMetric(name string, value int64) (err error) {
	if name == "fail" {
		err = errors.New("failed saved metric fail")
		return
	}

	return
}

func (m *mockStorage) GetAllMetrics() map[string]map[string]string {
	return nil
}

func (m *mockStorage) CheckDBConnect() error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockStorage) GetType() string {
	return "mock"
}

func (m *mockStorage) PutCounterMetrics(map[string]int64) (err error) {
	return nil
}

func (m *mockStorage) PutGaugeMetrics(map[string]float64) (err error) {
	return nil
}

func TestPutValue(t *testing.T) {
	storage := &mockStorage{}
	server := &HTTPServer{
		storage: storage,
	}

	tests := []struct {
		name   string
		mType  string
		mName  string
		mValue string
		stCode int
	}{
		{
			name:   "correct gauge metric",
			mType:  "gauge",
			mName:  "pass",
			mValue: "0",
			stCode: http.StatusOK,
		},
		{
			name:   "incorrect gauge metric",
			mType:  "gauge",
			mName:  "fail",
			mValue: "0",
			stCode: http.StatusInternalServerError,
		},
		{
			name:   "correct counter metric",
			mType:  "counter",
			mName:  "pass",
			mValue: "0",
			stCode: http.StatusOK,
		},
		{
			name:   "incorrect counter metric",
			mType:  "counter",
			mName:  "fail",
			mValue: "0",
			stCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test %s", test.name), func(t *testing.T) {
			ctx := context.WithValue(context.Background(), contextMetricType, test.mType)
			ctx = context.WithValue(ctx, contextMetricName, test.mName)
			ctx = context.WithValue(ctx, contextMetricValue, test.mValue)
			req := &http.Request{}
			req = req.WithContext(ctx)
			rw := httptest.NewRecorder()
			server.putValue(rw, req)
			assert.Equal(t, rw.Code, test.stCode)
		})
	}
}

func TestGetValue(t *testing.T) {
	storage := &mockStorage{}
	server := &HTTPServer{
		storage: storage,
	}

	tests := []struct {
		name   string
		mType  string
		mName  string
		stCode int
	}{
		{
			name:   "correct gauge metric",
			mType:  "gauge",
			mName:  "pass",
			stCode: http.StatusOK,
		},
		{
			name:   "incorrect gauge metric",
			mType:  "gauge",
			mName:  "fail",
			stCode: http.StatusNotFound,
		},
		{
			name:   "correct counter metric",
			mType:  "counter",
			mName:  "pass",
			stCode: http.StatusOK,
		},
		{
			name:   "incorrect counter metric",
			mType:  "counter",
			mName:  "fail",
			stCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test %s", test.name), func(t *testing.T) {
			ctx := context.WithValue(context.Background(), contextMetricType, test.mType)
			ctx = context.WithValue(ctx, contextMetricName, test.mName)
			req := &http.Request{}
			req = req.WithContext(ctx)
			rw := httptest.NewRecorder()
			server.getValue(rw, req)
			assert.Equal(t, rw.Code, test.stCode)
		})
	}
}

func TestCheckUpdateContext(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricValue string
		err         error
	}{
		{
			name:        "wrong metric type",
			metricType:  "test",
			metricValue: "1",
			err:         errors.New("unknown metric type"),
		},
		{
			name:        "wrong gauge value",
			metricType:  "gauge",
			metricValue: "gauge",
			err:         errors.New("wrong gauge value"),
		},
		{
			name:        "wrong counter value",
			metricType:  "counter",
			metricValue: "counter",
			err:         errors.New("wrong counter value"),
		},
		{
			name:        "float counter value",
			metricType:  "counter",
			metricValue: "1.1",
			err:         errors.New("wrong counter value"),
		},
		{
			name:        "correct gauge value",
			metricType:  "gauge",
			metricValue: "1.1",
			err:         nil,
		},
		{
			name:        "correct counter value",
			metricType:  "counter",
			metricValue: "1",
			err:         nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test %s", test.name), func(t *testing.T) {
			err := checkUpdateContext(test.metricType, test.metricValue)
			assert.Equal(t, err, test.err)
		})
	}
}

func TestCheckMetricName(t *testing.T) {
	tests := []struct {
		name       string
		metricType string
		metricName string
		err        error
	}{
		{
			name:       "wrong metric type",
			metricType: "test",
			metricName: "test",
			err:        errors.New("unknown metric type"),
		},
		{
			name:       "empty gauge metric",
			metricType: "gauge",
			metricName: "",
			err:        errors.New("empty metric name"),
		},
		{
			name:       "empty counter metric",
			metricType: "counter",
			metricName: "",
			err:        errors.New("empty metric name"),
		},
		{
			name:       "correct gauge metric",
			metricType: "gauge",
			metricName: "Alloc",
			err:        nil,
		},
		{
			name:       "correct counter metric",
			metricType: "counter",
			metricName: "PollCount",
			err:        nil,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test %s", test.name), func(t *testing.T) {
			err := checkMetricName(test.metricType, test.metricName)
			assert.Equal(t, err, test.err)
		})
	}
}

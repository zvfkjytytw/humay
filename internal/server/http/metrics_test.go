package humayhttpserver

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

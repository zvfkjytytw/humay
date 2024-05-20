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
			name:       "wrong gauge name",
			metricType: "gauge",
			metricName: "gauge",
			err:        errors.New("unknown metric name"),
		},
		{
			name:       "wrong counter name",
			metricType: "counter",
			metricName: "counter",
			err:        errors.New("unknown metric name"),
		},
		{
			name:       "float gauge name",
			metricType: "gauge",
			metricName: "Alloc",
			err:        nil,
		},
		{
			name:       "correct counter name",
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

// func TestHTTPServer_getValue(t *testing.T) {
// 	type fields struct {
// 		config  *HTTPConfig
// 		logger  *zap.Logger
// 		storage *metricsStorage.MemStorage
// 	}
// 	type args struct {
// 		w http.ResponseWriter
// 		r *http.Request
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			h := &HTTPServer{
// 				config:  tt.fields.config,
// 				logger:  tt.fields.logger,
// 				storage: tt.fields.storage,
// 			}
// 			h.getValue(tt.args.w, tt.args.r)
//  		})
// 	}
// }

package humaystorage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testStorage = NewStorage()

func TestGetType(t *testing.T) {
	sType := testStorage.GetType()
	assert.Equal(t, sType, "struct")
}

func TestGaugeMetric(t *testing.T) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.YearDay()

	tests := []struct {
		name  string
		value float64
	}{
		{
			name:  "A",
			value: 1.1,
		},
		{
			name:  "B",
			value: 2.2,
		},
	}
	for i := 1; i <= 5; i++ {
		for j, test := range tests {
			t.Run(fmt.Sprintf("Step_%d_Key_%d", i, j), func(t *testing.T) {
				testStorage.PutGaugeMetric(test.name, test.value)
				metric, err := testStorage.GetGaugeMetric(test.name)
				assert.NoError(t, err)
				assert.Equal(t, metric, test.value)

				metrics, err := testStorage.GetGaugeMetrics(test.name)
				assert.NoError(t, err)
				assert.Equal(t, i, len(metrics))
				for key, metric := range metrics {
					keyYear := time.Unix(key, 0).Year()
					keyMonth := time.Unix(key, 0).Month()
					keyDay := time.Unix(key, 0).YearDay()

					assert.Equal(t, metric, test.value)
					assert.Equal(t, keyYear, year)
					assert.Equal(t, keyMonth, month)
					assert.Equal(t, keyDay, day)
				}
			})
		}
		time.Sleep(time.Second)
	}
}

func TestCounterMetric(t *testing.T) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.YearDay()

	tests := []struct {
		name  string
		value int64
	}{
		{
			name:  "A",
			value: 1,
		},
		{
			name:  "B",
			value: 2,
		},
	}
	for i := 1; i <= 5; i++ {
		for j, test := range tests {
			t.Run(fmt.Sprintf("Step_%d_Key_%d", i, j), func(t *testing.T) {
				testStorage.PutCounterMetric(test.name, test.value)
				metric, err := testStorage.GetCounterMetric(test.name)
				assert.NoError(t, err)
				assert.Equal(t, metric, int64(i)*test.value)

				metrics, err := testStorage.GetCounterMetrics(test.name)
				assert.NoError(t, err)
				assert.Equal(t, i, len(metrics))
				for key, metric := range metrics {
					keyYear := time.Unix(key, 0).Year()
					keyMonth := time.Unix(key, 0).Month()
					keyDay := time.Unix(key, 0).YearDay()

					assert.Equal(t, metric, test.value)
					assert.Equal(t, keyYear, year)
					assert.Equal(t, keyMonth, month)
					assert.Equal(t, keyDay, day)
				}
			})
		}
		time.Sleep(time.Second)
	}
}

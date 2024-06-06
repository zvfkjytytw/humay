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
				err := testStorage.PutGaugeMetric(test.name, test.value)
				assert.NoError(t, err)
				metric, err := testStorage.GetGaugeMetric(test.name)
				assert.NoError(t, err)
				assert.Equal(t, metric, test.value)
			})
		}
		time.Sleep(time.Second)
	}
}

func TestCounterMetric(t *testing.T) {
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
				err := testStorage.PutCounterMetric(test.name, test.value)
				assert.NoError(t, err)
				metric, err := testStorage.GetCounterMetric(test.name)
				assert.NoError(t, err)
				assert.Equal(t, metric, int64(i)*test.value)
			})
		}
		time.Sleep(time.Second)
	}
}

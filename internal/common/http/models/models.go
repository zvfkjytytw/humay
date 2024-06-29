package httpmodels

var (
	CounterMetric      = "counter"
	GaugeMetric        = "gauge"
	MetricTypes        = []string{CounterMetric, GaugeMetric}
	UpdateHandler      = "/update"
	ValueHandler       = "/value"
	UpdateHandlerSlash = "/update/"
	ValueHandlerSlash  = "/value/"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

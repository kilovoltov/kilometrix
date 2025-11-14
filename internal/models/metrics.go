package models

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

// Metric — описание метрики
type Metric struct {
	Name  string
	Type  MetricType
	Value string
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type Storage map[string]*Metric

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
// type Metrics struct {
//     ID    string   `json:"id"`
//     MType string   `json:"type"`
//     Delta *int64   `json:"delta,omitempty"`
//     Value *float64 `json:"value,omitempty"`
//     Hash  string   `json:"hash,omitempty"`
// }

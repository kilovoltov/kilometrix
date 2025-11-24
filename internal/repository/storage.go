package repository

import (
	"fmt"

	"github.com/kilovoltov/kilometrix/internal/models"
)

// Интерфейс для работы с хранилищем метрик
type MetricStorage interface {
	AddGauge(name string, value float64) error
	AddCounter(name string, value int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetGaugesNames() []string
	GetCounterNames() []string
	Snapshot() []models.Metrics
}

// Тип для хранения метрик в памяти
type MemStorage struct {
	metricsGauge   map[string]float64
	metricsCounter map[string]int64
}

// Конструктор MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metricsGauge:   make(map[string]float64),
		metricsCounter: make(map[string]int64),
	}
}

func (ms *MemStorage) GetGaugesNames() []string {
    gKeys := make([]string, len(ms.metricsGauge))
    i := 0
    for k := range ms.metricsGauge {
        gKeys[i] = k
        i++
    }
    return gKeys
}

func (ms *MemStorage) GetCounterNames() []string {
    cKeys := make([]string, len(ms.metricsCounter))
    i := 0
    for k := range ms.metricsCounter {
        cKeys[i] = k
        i++
    }
    return cKeys
}

// Добавление метрики типа Gauge в хранилище
func (ms *MemStorage) AddGauge(name string, value float64) error {
	ms.metricsGauge[name] = value
	return nil
}

// Добавление метрики типа Counter в хранилище
func (ms *MemStorage) AddCounter(name string, value int64) error {
	ms.metricsCounter[name] += value
	return nil
}

// Получение метрики по имени
func (ms *MemStorage) GetGauge(name string) (float64, error) {
	metric, ok := ms.metricsGauge[name]
	if !ok {
		return 0, fmt.Errorf("metric %s not found", name)
	}
	return metric, nil
}

func (ms *MemStorage) GetCounter(name string) (int64, error) {
	metric, ok := ms.metricsCounter[name]
	if !ok {
		return 0, fmt.Errorf("metric %s not found", name)
	}
	return metric, nil
}

func (ms *MemStorage) Snapshot() []models.Metrics {
	snapshot := make([]models.Metrics, 0, len(ms.metricsCounter)+len(ms.metricsGauge))

	for _, gName := range ms.GetGaugesNames() {
		v, _ := ms.GetGauge(gName)
		snapshot = append(snapshot, models.Metrics{ID: gName, MType: "gauge", Delta: nil, Value: &v})
	}
	for _, cName := range ms.GetCounterNames() {
		d, _ := ms.GetCounter(cName)
		snapshot = append(snapshot, models.Metrics{ID: cName, MType: "counter", Delta: &d, Value: nil})
	}

	return snapshot
}

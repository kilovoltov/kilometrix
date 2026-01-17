// Package utils provides some useful funcs
package utils

import(
	"github.com/kilovoltov/kilometrix/internal/models"
)

// RemoveDuplicates убирает повторяющиеся элементы, оставляя
// последний по уникалной паре (ID + MType)
// если тип counter то значения суммируются
func RemoveDuplicates(metrics []models.Metrics) []models.Metrics {
	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for _, m := range metrics {
		if m.MType == "gauge" && m.Value != nil {
			gauges[m.ID] = *m.Value
		} else if m.MType == "counter" && m.Delta != nil {
			counters[m.ID] += *m.Delta
		}
	}

	var result []models.Metrics
	for id, val := range gauges {
		result = append(result, models.Metrics{
			ID:    id,
			MType: "gauge",
			Value: &val,
		})
	}
	for id, delta := range counters {
		result = append(result, models.Metrics{
			ID:    id,
			MType: "counter",
			Delta: &delta,
		})
	}
	return result
}
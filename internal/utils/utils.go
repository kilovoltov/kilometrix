// Package utils provides some useful funcs
package utils

import(
	"github.com/kilovoltov/kilometrix/internal/models"
)

// RemoveDuplicatesLast убирает повторяющиеся элементы, оставляя
// последний по уникалной паре (ID + MType)
func RemoveDuplicatesLast(metrics []models.Metrics) []models.Metrics {
	type metricKey struct {
		ID    string
		MType string
	}
	seen := make(map[metricKey]bool)
	var result []models.Metrics

	for i := len(metrics) - 1; i >= 0; i-- {
		key := metricKey{ID: metrics[i].ID, MType: metrics[i].MType}
		if !seen[key] {
			seen[key] = true
			result = append(result, metrics[i])
		}
	}

	// Разворачиваем результат в исходный порядок
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
package sender

import (
	"encoding/json"
	"fmt"
	"strconv"
	"errors"

	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/go-resty/resty/v2"
)

// SendMetricsJSON отправляет несколько метрик за один раз
func SendMetricsJSON(serverAddress string, client *resty.Client, metrics models.Storage, key string) error {
	url := fmt.Sprintf("http://%s/updates/",
		serverAddress,
	)
	data := make([]models.Metrics, 0, len(metrics))

	for _, metric := range metrics {
		switch metric.Type {
		case "gauge":
			valueFloat, _ := strconv.ParseFloat(metric.Value, 64)
			data = append(data, models.Metrics{
				ID:    metric.Name,
				MType: string(metric.Type),
				Value: &valueFloat,
			})
		case "counter":
			valueInt, _ := strconv.ParseInt(metric.Value, 10, 64)
			data = append(data, models.Metrics{
				ID:    metric.Name,
				MType: string(metric.Type),
				Delta: &valueInt,
			})
		}
	}
	// Сериализуем структуру в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Ошибка сериализации в JSON: %s\n", err)
		return err
	}

	// Генерируем хэш для jsonData
	hash := HashCount(jsonData, key)

	// Сжимаем JSON с помощью gzip
	compressedData, err := GzipCompress(jsonData)
	if err != nil {
		fmt.Printf("Ошибка сжатия gzip: %s\n", err)
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetHeader("HashSHA256", hash).
		SetBody(compressedData).
		Post(url)

	if err != nil {
		fmt.Println("failed to send metrics")
		return err
	}
	if resp.StatusCode() != 200 {
		fmt.Println("non-200 status for metrics")
		return errors.New(resp.Status())
	}

	return nil
}

// SendMetricJSON отправляет одну метрику
func SendMetricJSON(serverAddress string, client *resty.Client, metric models.Metric) error {
	url := fmt.Sprintf("http://%s/update",
		serverAddress,
	)
	var data models.Metrics

	switch metric.Type {
	case "gauge":
		valueFloat, _ := strconv.ParseFloat(metric.Value, 64)
		data = models.Metrics{
			ID:    metric.Name,
			MType: string(metric.Type),
			Value: &valueFloat,
		}
	case "counter":
		valueInt, _ := strconv.ParseInt(metric.Value, 10, 64)
		data = models.Metrics{
			ID:    metric.Name,
			MType: string(metric.Type),
			Delta: &valueInt,
		}
	}

	// Сериализуем структуру в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Ошибка сериализации в JSON: %s\n", err)
		return err
	}

	// Сжимаем JSON с помощью gzip
	compressedData, err := GzipCompress(jsonData)
	if err != nil {
		fmt.Printf("Ошибка сжатия gzip: %s\n", err)
		return err
	}

	resp, err := client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(compressedData).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send metric %s: %w", metric.Name, err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("non-200 status for metric %s: %s", metric.Name, resp.Status())
	}

	return nil
}

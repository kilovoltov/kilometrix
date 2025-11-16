package sender

import (
	"fmt"
	"strconv"
	"encoding/json"
	"log"

	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/go-resty/resty/v2"
)

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
		log.Fatal("Ошибка сериализации в JSON:", err)
	}

	// Сжимаем JSON с помощью gzip
	compressedData, _ := GzipCompress(jsonData)

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
	// else {
	// 	fmt.Println(string(resp.Body()))
	// }
	return nil
}

package sender

import (
	"fmt"
	"strconv"

	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/go-resty/resty/v2"
)

// отправляет одну метрику на сервер
func SendMetric(serverAddress string, client *resty.Client, metric models.Metric) error {
	url := fmt.Sprintf("http://%s/update/%s/%s/%s",
		serverAddress,
		metric.Type,
		metric.Name,
		metric.Value,
	)

	resp, err := client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send metric %s: %w", metric.Name, err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("non-200 status for metric %s: %s", metric.Name, resp.Status())
	}
	return nil
}

func SendMetricJson(serverAddress string, client *resty.Client, metric models.Metric) error {
	url := fmt.Sprintf("http://%s/update",
		serverAddress,
	)
	var data models.Metrics

	if metric.Type == "gauge" {
		valueFloat, _ := strconv.ParseFloat(metric.Value, 64)
		data = models.Metrics{
			ID:    metric.Name,
			MType: string(metric.Type),
			Value: &valueFloat,
		}
	} else {
		valueInt, _ := strconv.ParseInt(metric.Value, 10, 64)
		data = models.Metrics{
			ID:    metric.Name,
			MType: string(metric.Type),
			Delta: &valueInt,
		}
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send metric %s: %w", metric.Name, err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("non-200 status for metric %s: %s", metric.Name, resp.Status())
	} else {
		fmt.Println(string(resp.Body()))
	}
	return nil
}

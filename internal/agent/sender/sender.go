package sender

import (
	"fmt"

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

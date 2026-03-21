package sender

import (
	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/go-resty/resty/v2"
)

// Task представляет задачу для воркера по отправке метрик
type Task struct {
	Metrics models.Storage // Метрики для отправки
	Server  string         // Адрес сервера
	Client  *resty.Client  // HTTP клиент
	Secret  string         // Секретный ключ для хеширования
}

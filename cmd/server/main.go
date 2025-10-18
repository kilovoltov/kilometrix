package main

import (
    "fmt"
    "net/http"
    "strconv"
    "strings"
)

// Интерфейс для работы с хранилищем метрик
type MetricStorage interface {
    AddGauge(name string, value float64) error
    AddCounter(name string, value int64) error
    GetGauge(name string) (float64, error)
    GetCounter(name string) (int64, error)
}

// Тип для хранения метрик в памяти
type MemStorage struct {
    metricsGauge map[string]float64
    metricsCounter map[string]int64
}

// Конструктор MemStorage
func NewMemStorage() *MemStorage {
    return &MemStorage{
        metricsGauge: make(map[string]float64),
        metricsCounter: make(map[string]int64),
    }
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

// Функция для обработки запросов на добавление метрик
func handleMetricUpdate(w http.ResponseWriter, r *http.Request) {
    // Разбираем URL
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) != 5 {
        http.Error(w, "Invalid URL structure", http.StatusBadRequest)
        return
    }

    metricType, metricName, metricValue := parts[2], parts[3], parts[4]

    // Преобразуем значение в нужный формат
    var e error
    switch metricType {
    case "gauge":
        value, err := strconv.ParseFloat(metricValue, 64)
        if err != nil {
            http.Error(w, "Invalid gauge value", http.StatusBadRequest)
            return
        }
        e = storage.AddGauge(metricName, value)
    case "counter":
        value, err := strconv.ParseInt(metricValue, 10, 64)
        if err != nil {
            http.Error(w, "Invalid counter value", http.StatusBadRequest)
            return
        }
        e = storage.AddCounter(metricName, value)
    default:
        http.Error(w, "Invalid metric type", http.StatusBadRequest)
        return
    }

    // Если произошла ошибка при добавлении
    if e != nil {
        http.Error(w, e.Error(), http.StatusInternalServerError)
        return
    }

    // Отправляем успешный ответ
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Metric %s updated successfully", metricName)
}

// Функция для обработки запроса на получение метрики
func handleMetricGet(w http.ResponseWriter, r *http.Request) {
    // Разбираем URL
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) != 4 {
        http.Error(w, "Invalid URL structure", http.StatusBadRequest)
        return
    }

    metricType, metricName := parts[2], parts[3]
    var vString string
    var e error
    switch metricType {
    case "gauge":
        value, _ := storage.GetGauge(metricName)
        vString = strconv.FormatFloat(value, 'f', -1, 64)
    case "counter":
        value, _ := storage.GetCounter(metricName)
        vString = strconv.FormatInt(value, 10)
    default:
        http.Error(w, "Invalid metric type", http.StatusBadRequest)
        return
    }

    // Если произошла ошибка при добавлении
    if e != nil {
        http.Error(w, e.Error(), http.StatusBadRequest)
        return
    }

    w.Write([]byte(vString))
}

// Глобальная переменная для хранилища метрик
var storage = NewMemStorage()

func main() {
    http.HandleFunc("/update/", handleMetricUpdate)
    http.HandleFunc("/get/", handleMetricGet)

    fmt.Println("Server started at http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        panic(err)
    }
}
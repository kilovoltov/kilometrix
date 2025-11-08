package handlers

import (
	"fmt"
	"strconv"
	"net/http"
	"html/template"
	"sort"
	"github.com/kilovoltov/kilometrix/internal/repository"

	"github.com/go-chi/chi/v5"
)

// Тип для хранения метрик в памяти
type Stor struct {
    repo repository.MetricStorage
}

// Конструктор Stor
func NewStor(rs repository.MetricStorage) *Stor {
    return &Stor{
        repo: rs,
    }
}

// Функция для обработки запросов на добавление метрик
func (s *Stor) HandleMetricUpdate(w http.ResponseWriter, r *http.Request) {

    metricType := chi.URLParam(r, "metricType")
    metricName := chi.URLParam(r, "metricName")
    metricValue := chi.URLParam(r, "metricValue")

    if metricName == "" {
        http.Error(w, "Invalid metric name", http.StatusNotFound)
        return
    }
    if metricValue == "" {
        http.Error(w, "Invalid metric value", http.StatusBadRequest)
        return
    }
    // Преобразуем значение в нужный формат
    var e error
    switch metricType {
    case "gauge":
        value, err := strconv.ParseFloat(metricValue, 64)
        if err != nil {
            http.Error(w, "Invalid gauge value", http.StatusBadRequest)
            return
        }
        e = s.repo.AddGauge(metricName, value)
    case "counter":
        value, err := strconv.ParseInt(metricValue, 10, 64)
        if err != nil {
            http.Error(w, "Invalid counter value", http.StatusBadRequest)
            return
        }
        e = s.repo.AddCounter(metricName, value)
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
func (s *Stor) HandleMetricGet(w http.ResponseWriter, r *http.Request) {
    metricType, metricName := chi.URLParam(r, "metricType"), chi.URLParam(r, "metricName")
    var vString string
    var e error
    switch metricType {
    case "gauge":
        var value float64
        value, e = s.repo.GetGauge(metricName)
        vString = strconv.FormatFloat(value, 'f', -1, 64)
    case "counter":
        var value int64
        value, e = s.repo.GetCounter(metricName)
        vString = strconv.FormatInt(value, 10)
    default:
        http.Error(w, "Invalid metric type", http.StatusBadRequest)
        return
    }

    if e != nil {
        http.Error(w, e.Error(), http.StatusNotFound)
        return
    }

    w.Write([]byte(vString))
}

func (s *Stor) HandleMain(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "Invalid path", http.StatusNotFound)
        return
    }

    tmpl, err := template.ParseFiles("index.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    metricList := make(map[string]string)
    var gValue float64
    var cValue int64

    for _, gName := range s.repo.GetGaugesNames() {
        gValue, _ = s.repo.GetGauge(gName)
        metricList[gName] = strconv.FormatFloat(gValue, 'f', -1, 64)
    }
    for _, cName := range s.repo.GetCounterNames() {
        cValue, _ = s.repo.GetCounter(cName)
        metricList[cName] = strconv.FormatInt(cValue, 10)
    }

    type KeyValue struct {
        Key   string
        Value string
    }

    // Сортируем ключи
    keys := make([]string, 0, len(metricList))
    for k := range metricList {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    // Преобразуем в срез
    var pairs []KeyValue
    for _, k := range keys {
        pairs = append(pairs, KeyValue{Key: k, Value: metricList[k]})
    }

    tmpl.Execute(w, pairs)
}
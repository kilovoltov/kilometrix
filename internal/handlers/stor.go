package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/kilovoltov/kilometrix/internal/models"
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
	// w.WriteHeader(http.StatusOK)
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

// Функция для отображения списка метрик
func (s *Stor) HandleMain(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Invalid path", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("index.html") // Шаблон
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, pairs)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		return
	}
}

func (s *Stor) HandleMetricUpdateJSON(w http.ResponseWriter, r *http.Request) {
	var metricsJSON models.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// десериализуем JSON в metricsJSON
	if err = json.Unmarshal(buf.Bytes(), &metricsJSON); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Преобразуем значение в нужный формат
	switch metricsJSON.MType {
	case "gauge":
		s.repo.AddGauge(metricsJSON.ID, *metricsJSON.Value)
	case "counter":
		s.repo.AddCounter(metricsJSON.ID, *metricsJSON.Delta)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (s *Stor) HandleValueJSON(w http.ResponseWriter, r *http.Request) {
	var metricsJSON models.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// десериализуем JSON в metricsJSON
	if err = json.Unmarshal(buf.Bytes(), &metricsJSON); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Преобразуем значение в нужный формат
	var data models.Metrics
	switch metricsJSON.MType {
	case "gauge":
		gValue, _ := s.repo.GetGauge(metricsJSON.ID)
		data = models.Metrics{
			ID:    metricsJSON.ID,
			MType: metricsJSON.MType,
			Value: &gValue,
		}
	case "counter":
		cValue, _ := s.repo.GetCounter(metricsJSON.ID)
		data = models.Metrics{
			ID:    metricsJSON.ID,
			MType: metricsJSON.MType,
			Delta: &cValue,
		}
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dataJSON)
}

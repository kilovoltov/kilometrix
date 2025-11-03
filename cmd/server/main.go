package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// Интерфейс для работы с хранилищем метрик
type MetricStorage interface {
	AddGauge(name string, value float64) error
	AddCounter(name string, value int64) error
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetGaugesNames() []string
	GetCounterNames() []string
}

// Тип для хранения метрик в памяти
type MemStorage struct {
	metricsGauge   map[string]float64
	metricsCounter map[string]int64
}

// Конструктор MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metricsGauge:   make(map[string]float64),
		metricsCounter: make(map[string]int64),
	}
}

func (ms *MemStorage) GetGaugesNames() []string {
	gKeys := make([]string, len(ms.metricsGauge))
	i := 0
	for k := range ms.metricsGauge {
		gKeys[i] = k
		i++
	}
	return gKeys
}

func (ms *MemStorage) GetCounterNames() []string {
	cKeys := make([]string, len(ms.metricsCounter))
	i := 0
	for k := range ms.metricsCounter {
		cKeys[i] = k
		i++
	}
	return cKeys
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
	metricType, metricName := chi.URLParam(r, "metricType"), chi.URLParam(r, "metricName")
	var vString string
	var e error
	switch metricType {
	case "gauge":
		var value float64
		value, e = storage.GetGauge(metricName)
		vString = strconv.FormatFloat(value, 'f', -1, 64)
	case "counter":
		var value int64
		value, e = storage.GetCounter(metricName)
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

func handleMain(w http.ResponseWriter, r *http.Request) {
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

	for _, gName := range storage.GetGaugesNames() {
		gValue, _ = storage.GetGauge(gName)
		metricList[gName] = strconv.FormatFloat(gValue, 'f', -1, 64)
	}
	for _, cName := range storage.GetCounterNames() {
		cValue, _ = storage.GetCounter(cName)
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

// Глобальная переменная для хранилища метрик
var storage = NewMemStorage()

func main() {
	// переменная для адреса сервера со значением по умолчанию
	var addr = flag.String("a", "localhost:8080", "address of the server")

	flag.Parse()

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", handleMetricUpdate)
	r.Get("/value/{metricType}/{metricName}", handleMetricGet)
	r.Get("/", handleMain)

	fmt.Printf("Server started at http://%s\n", *addr)
	err := http.ListenAndServe(*addr, r)
	if err != nil {
		panic(err)
	}
}

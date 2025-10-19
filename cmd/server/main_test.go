package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Тест для конструктора NewMemStorage
func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	
	if storage == nil {
		t.Fatal("NewMemStorage() вернул nil")
	}
	
	if storage.metricsGauge == nil {
		t.Error("metricsGauge не инициализирован")
	}
	
	if storage.metricsCounter == nil {
		t.Error("metricsCounter не инициализирован")
	}
	
	if len(storage.metricsGauge) != 0 {
		t.Error("metricsGauge не пустой после создания")
	}
	
	if len(storage.metricsCounter) != 0 {
		t.Error("metricsCounter не пустой после создания")
	}
}

// Тесты для MemStorage.AddGauge
func TestMemStorage_AddGauge(t *testing.T) {
	storage := NewMemStorage()
	
	// Тест добавления gauge метрики
	err := storage.AddGauge("test_gauge", 123.45)
	if err != nil {
		t.Errorf("AddGauge() вернул ошибку: %v", err)
	}
	
	// Проверяем, что метрика добавлена
	value, err := storage.GetGauge("test_gauge")
	if err != nil {
		t.Errorf("GetGauge() вернул ошибку: %v", err)
	}
	
	if value != 123.45 {
		t.Errorf("Ожидаемое значение 123.45, получено %f", value)
	}
	
	// Тест перезаписи gauge метрики
	err = storage.AddGauge("test_gauge", 999.99)
	if err != nil {
		t.Errorf("AddGauge() при перезаписи вернул ошибку: %v", err)
	}
	
	value, err = storage.GetGauge("test_gauge")
	if err != nil {
		t.Errorf("GetGauge() после перезаписи вернул ошибку: %v", err)
	}
	
	if value != 999.99 {
		t.Errorf("Ожидаемое значение после перезаписи 999.99, получено %f", value)
	}
}

// Тесты для MemStorage.AddCounter
func TestMemStorage_AddCounter(t *testing.T) {
	storage := NewMemStorage()
	
	// Тест добавления counter метрики
	err := storage.AddCounter("test_counter", 10)
	if err != nil {
		t.Errorf("AddCounter() вернул ошибку: %v", err)
	}
	
	// Проверяем, что метрика добавлена
	value, err := storage.GetCounter("test_counter")
	if err != nil {
		t.Errorf("GetCounter() вернул ошибку: %v", err)
	}
	
	if value != 10 {
		t.Errorf("Ожидаемое значение 10, получено %d", value)
	}
	
	// Тест добавления к существующей counter метрике
	err = storage.AddCounter("test_counter", 5)
	if err != nil {
		t.Errorf("AddCounter() при добавлении к существующей вернул ошибку: %v", err)
	}
	
	value, err = storage.GetCounter("test_counter")
	if err != nil {
		t.Errorf("GetCounter() после добавления вернул ошибку: %v", err)
	}
	
	if value != 15 {
		t.Errorf("Ожидаемое значение после добавления 15, получено %d", value)
	}
}

// Тесты для MemStorage.GetGauge
func TestMemStorage_GetGauge(t *testing.T) {
	storage := NewMemStorage()
	
	// Тест получения несуществующей gauge метрики
	_, err := storage.GetGauge("nonexistent")
	if err == nil {
		t.Error("GetGauge() для несуществующей метрики должен вернуть ошибку")
	}
	
	expectedError := "metric nonexistent not found"
	if err.Error() != expectedError {
		t.Errorf("Ожидаемая ошибка '%s', получено '%s'", expectedError, err.Error())
	}
	
	// Тест получения существующей gauge метрики
	storage.AddGauge("existing", 42.0)
	value, err := storage.GetGauge("existing")
	if err != nil {
		t.Errorf("GetGauge() для существующей метрики вернул ошибку: %v", err)
	}
	
	if value != 42.0 {
		t.Errorf("Ожидаемое значение 42.0, получено %f", value)
	}
}

// Тесты для MemStorage.GetCounter
func TestMemStorage_GetCounter(t *testing.T) {
	storage := NewMemStorage()
	
	// Тест получения несуществующей counter метрики
	_, err := storage.GetCounter("nonexistent")
	if err == nil {
		t.Error("GetCounter() для несуществующей метрики должен вернуть ошибку")
	}
	
	expectedError := "metric nonexistent not found"
	if err.Error() != expectedError {
		t.Errorf("Ожидаемая ошибка '%s', получено '%s'", expectedError, err.Error())
	}
	
	// Тест получения существующей counter метрики
	storage.AddCounter("existing", 100)
	value, err := storage.GetCounter("existing")
	if err != nil {
		t.Errorf("GetCounter() для существующей метрики вернул ошибку: %v", err)
	}
	
	if value != 100 {
		t.Errorf("Ожидаемое значение 100, получено %d", value)
	}
}

// Вспомогательная функция для создания HTTP запроса
func createTestRequest(method, url string, body string) *http.Request {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, url, bodyReader)
	return req
}

// Тесты для handleMetricUpdate
func TestHandleMetricUpdate(t *testing.T) {
	// Создаем временное хранилище для тестов
	originalStorage := storage
	defer func() { storage = originalStorage }()
	
	storage = NewMemStorage()
	
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid gauge metric",
			url:            "/update/gauge/test_gauge/123.45",
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric test_gauge updated successfully",
		},
		{
			name:           "Valid counter metric",
			url:            "/update/counter/test_counter/42",
			expectedStatus: http.StatusOK,
			expectedBody:   "Metric test_counter updated successfully",
		},
		{
			name:           "Invalid URL structure - too few parts",
			url:            "/update/gauge/test",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Invalid URL structure",
		},
		{
			name:           "Invalid metric name - empty",
			url:            "/update/gauge//123.45",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Invalid metric name",
		},
		{
			name:           "Invalid metric type",
			url:            "/update/invalid/test/123",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type",
		},
		{
			name:           "Invalid gauge value",
			url:            "/update/gauge/test/invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid gauge value",
		},
		{
			name:           "Invalid counter value",
			url:            "/update/counter/test/invalid",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid counter value",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("POST", tt.url, "")
			w := httptest.NewRecorder()
			
			handleMetricUpdate(w, req)
			
			if w.Code != tt.expectedStatus {
				t.Errorf("Ожидаемый статус %d, получен %d", tt.expectedStatus, w.Code)
			}
			
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Ожидаемое тело ответа содержит '%s', получено '%s'", 
					tt.expectedBody, w.Body.String())
			}
		})
	}
}

// Тесты для handleMetricGet
func TestHandleMetricGet(t *testing.T) {
	// Создаем временное хранилище для тестов
	originalStorage := storage
	defer func() { storage = originalStorage }()
	
	storage = NewMemStorage()
	
	// Добавляем тестовые метрики
	storage.AddGauge("test_gauge", 123.45)
	storage.AddCounter("test_counter", 42)
	
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid gauge metric",
			url:            "/value/gauge/test_gauge",
			expectedStatus: http.StatusOK,
			expectedBody:   "123.45",
		},
		{
			name:           "Valid counter metric",
			url:            "/value/counter/test_counter",
			expectedStatus: http.StatusOK,
			expectedBody:   "42",
		},
		{
			name:           "Invalid URL structure - too few parts",
			url:            "/value/gauge",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL structure",
		},
		{
			name:           "Invalid metric type",
			url:            "/value/invalid/test",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type",
		},
		{
			name:           "Non-existent gauge metric",
			url:            "/value/gauge/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "metric nonexistent not found",
		},
		{
			name:           "Non-existent counter metric",
			url:            "/value/counter/nonexistent",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "metric nonexistent not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest("GET", tt.url, "")
			w := httptest.NewRecorder()
			
			handleMetricGet(w, req)
			
			if w.Code != tt.expectedStatus {
				t.Errorf("Ожидаемый статус %d, получен %d", tt.expectedStatus, w.Code)
			}
			
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Ожидаемое тело ответа содержит '%s', получено '%s'", 
					tt.expectedBody, w.Body.String())
			}
		})
	}
}

// Интеграционный тест: полный цикл добавления и получения метрик
func TestMetricIntegration(t *testing.T) {
	// Создаем временное хранилище для тестов
	originalStorage := storage
	defer func() { storage = originalStorage }()
	
	storage = NewMemStorage()
	
	// Тест для gauge метрики
	t.Run("Gauge metric lifecycle", func(t *testing.T) {
		// Добавляем gauge метрику
		updateReq := createTestRequest("POST", "/update/gauge/cpu_usage/75.5", "")
		updateW := httptest.NewRecorder()
		handleMetricUpdate(updateW, updateReq)
		
		if updateW.Code != http.StatusOK {
			t.Errorf("Ошибка при добавлении gauge метрики, статус: %d", updateW.Code)
		}
		
		// Получаем gauge метрику
		getReq := createTestRequest("GET", "/value/gauge/cpu_usage", "")
		getW := httptest.NewRecorder()
		handleMetricGet(getW, getReq)
		
		if getW.Code != http.StatusOK {
			t.Errorf("Ошибка при получении gauge метрики, статус: %d", getW.Code)
		}
		
		if getW.Body.String() != "75.5" {
			t.Errorf("Ожидаемое значение '75.5', получено '%s'", getW.Body.String())
		}
	})
	
	// Тест для counter метрики
	t.Run("Counter metric lifecycle", func(t *testing.T) {
		// Добавляем counter метрику
		updateReq := createTestRequest("POST", "/update/counter/requests_total/10", "")
		updateW := httptest.NewRecorder()
		handleMetricUpdate(updateW, updateReq)
		
		if updateW.Code != http.StatusOK {
			t.Errorf("Ошибка при добавлении counter метрики, статус: %d", updateW.Code)
		}
		
		// Добавляем еще раз ту же counter метрику
		updateReq2 := createTestRequest("POST", "/update/counter/requests_total/5", "")
		updateW2 := httptest.NewRecorder()
		handleMetricUpdate(updateW2, updateReq2)
		
		if updateW2.Code != http.StatusOK {
			t.Errorf("Ошибка при повторном добавлении counter метрики, статус: %d", updateW2.Code)
		}
		
		// Получаем counter метрику
		getReq := createTestRequest("GET", "/value/counter/requests_total", "")
		getW := httptest.NewRecorder()
		handleMetricGet(getW, getReq)
		
		if getW.Code != http.StatusOK {
			t.Errorf("Ошибка при получении counter метрики, статус: %d", getW.Code)
		}
		
		if getW.Body.String() != "15" {
			t.Errorf("Ожидаемое значение '15', получено '%s'", getW.Body.String())
		}
	})
}

// Бенчмарк для MemStorage операций
func BenchmarkMemStorage_AddGauge(b *testing.B) {
	storage := NewMemStorage()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.AddGauge(fmt.Sprintf("gauge_%d", i), float64(i))
	}
}

func BenchmarkMemStorage_AddCounter(b *testing.B) {
	storage := NewMemStorage()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.AddCounter(fmt.Sprintf("counter_%d", i), int64(i))
	}
}

func BenchmarkMemStorage_GetGauge(b *testing.B) {
	storage := NewMemStorage()
	// Предварительно добавляем метрики
	for i := 0; i < 1000; i++ {
		storage.AddGauge(fmt.Sprintf("gauge_%d", i), float64(i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.GetGauge(fmt.Sprintf("gauge_%d", i%1000))
	}
}

func BenchmarkMemStorage_GetCounter(b *testing.B) {
	storage := NewMemStorage()
	// Предварительно добавляем метрики
	for i := 0; i < 1000; i++ {
		storage.AddCounter(fmt.Sprintf("counter_%d", i), int64(i))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.GetCounter(fmt.Sprintf("counter_%d", i%1000))
	}
}
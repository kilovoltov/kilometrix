package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/go-chi/chi/v5"
)

// MetricNotFoundError - кастомная ошибка для отсутствующих метрик
type MetricNotFoundError string

func (e MetricNotFoundError) Error() string {
	return string(e)
}

// MockStorage - мок-реализация MetricStorage для тестирования
type MockStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *MockStorage) CheckStorage() error {
	return nil
}

func (m *MockStorage) InitStorage() error {
	return nil
}

func (m *MockStorage) CloseStorage() error {
	return nil
}

func (m *MockStorage) AddGauge(name string, value float64) error {
	m.gauges[name] = value
	return nil
}

func (m *MockStorage) AddCounter(name string, value int64) error {
	m.counters[name] += value
	return nil
}

func (m *MockStorage) GetGauge(name string) (float64, error) {
	value, ok := m.gauges[name]
	if !ok {
		return 0, MetricNotFoundError("metric " + name + " not found")
	}
	return value, nil
}

func (m *MockStorage) GetCounter(name string) (int64, error) {
	value, ok := m.counters[name]
	if !ok {
		return 0, MetricNotFoundError("metric " + name + " not found")
	}
	return value, nil
}

func (m *MockStorage) GetGaugesNames() []string {
	gKeys := make([]string, len(m.gauges))
    i := 0
    for k := range m.gauges {
        gKeys[i] = k
        i++
    }
    return gKeys
}

func (m *MockStorage) GetCounterNames() []string {
	cKeys := make([]string, len(m.counters))
    i := 0
    for k := range m.counters {
        cKeys[i] = k
        i++
    }
    return cKeys
}

func (m *MockStorage) Snapshot() []models.Metrics {
	snapshot := make([]models.Metrics, 0, len(m.counters)+len(m.gauges))

	for _, gName := range m.GetGaugesNames() {
		v, _ := m.GetGauge(gName)
		snapshot = append(snapshot, models.Metrics{ID: gName, MType: "gauge", Delta: nil, Value: &v})
	}
	for _, cName := range m.GetCounterNames() {
		d, _ := m.GetCounter(cName)
		snapshot = append(snapshot, models.Metrics{ID: cName, MType: "counter", Delta: &d, Value: nil})
	}

	return snapshot
}

// Создаем контекст с роутингом
func contextWithRouteContext(ctx context.Context, rctx *chi.Context) context.Context {
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

func TestStor_HandleMetricUpdate(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		setupStorage func(*MockStorage)
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Успешное добавление gauge метрики",
			url:          "/update/gauge/testGauge/123.45",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusOK,
			expectedBody: "Metric testGauge updated successfully",
		},
		{
			name:         "Успешное добавление counter метрики",
			url:          "/update/counter/testCounter/100",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusOK,
			expectedBody: "Metric testCounter updated successfully",
		},
		{
			name:         "Пустое имя метрики",
			url:          "/update/gauge//123.45",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "Пустое значение метрики",
			url:          "/update/gauge/testGauge/",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "Неверный тип метрики",
			url:          "/update/invalid/testMetric/123",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "Неверное значение для gauge",
			url:          "/update/gauge/testGauge/invalid",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "Неверное значение для counter",
			url:          "/update/counter/testCounter/invalid",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage()
			tt.setupStorage(mockStorage)

			stor := NewStor(mockStorage)

			// Создаем тестовый запрос
			req := httptest.NewRequest("POST", tt.url, nil)

			// Настраиваем параметры роута
			rctx := chi.NewRouteContext()
			// Парсим URL для извлечения параметров
			parts := strings.Split(strings.TrimPrefix(tt.url, "/update/"), "/")
			if len(parts) >= 1 {
				rctx.URLParams.Add("metricType", parts[0])
			}
			if len(parts) >= 2 {
				rctx.URLParams.Add("metricName", parts[1])
			}
			if len(parts) >= 3 {
				rctx.URLParams.Add("metricValue", parts[2])
			}

			*req = *req.WithContext(contextWithRouteContext(req.Context(), rctx))

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			stor.HandleMetricUpdate(rr, req)

			// Проверяем код ответа
			if rr.Code != tt.expectedCode {
				t.Errorf("HandleMetricUpdate() код ответа = %v, ожидается %v", rr.Code, tt.expectedCode)
			}

			// Проверяем тело ответа
			if tt.expectedBody != "" && !strings.Contains(rr.Body.String(), tt.expectedBody) {
				t.Errorf("HandleMetricUpdate() тело ответа = %v, ожидается содержащее %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestStor_HandleMetricGet(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		setupStorage func(*MockStorage)
		expectedCode int
		expectedBody string
	}{
		{
			name: "Успешное получение существующей gauge метрики",
			url:  "/value/gauge/testGauge",
			setupStorage: func(ms *MockStorage) {
				ms.AddGauge("testGauge", 123.45)
			},
			expectedCode: http.StatusOK,
			expectedBody: "123.45",
		},
		{
			name: "Успешное получение существующей counter метрики",
			url:  "/value/counter/testCounter",
			setupStorage: func(ms *MockStorage) {
				ms.AddCounter("testCounter", 100)
			},
			expectedCode: http.StatusOK,
			expectedBody: "100",
		},
		{
			name:         "Получение несуществующей gauge метрики",
			url:          "/value/gauge/nonexistent",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "Получение несуществующей counter метрики",
			url:          "/value/counter/nonexistent",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "Неверный тип метрики",
			url:          "/value/invalid/testMetric",
			setupStorage: func(ms *MockStorage) {},
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage()
			tt.setupStorage(mockStorage)

			stor := NewStor(mockStorage)

			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", tt.url, nil)

			// Настраиваем параметры роута
			rctx := chi.NewRouteContext()
			// Парсим URL для извлечения параметров
			parts := strings.Split(strings.TrimPrefix(tt.url, "/value/"), "/")
			if len(parts) >= 1 {
				rctx.URLParams.Add("metricType", parts[0])
			}
			if len(parts) >= 2 {
				rctx.URLParams.Add("metricName", parts[1])
			}

			*req = *req.WithContext(contextWithRouteContext(req.Context(), rctx))

			// Создаем ResponseRecorder
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			stor.HandleMetricGet(rr, req)

			// Проверяем код ответа
			if rr.Code != tt.expectedCode {
				t.Errorf("HandleMetricGet() код ответа = %v, ожидается %v", rr.Code, tt.expectedCode)
			}

			// Проверяем тело ответа
			if tt.expectedBody != "" && rr.Body.String() != tt.expectedBody {
				t.Errorf("HandleMetricGet() тело ответа = %v, ожидается %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

// func TestStor_HandleMain(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		url          string
// 		setupStorage func(*MockStorage)
// 		expectedCode int
// 	}{
// 		{
// 			name:         "Успешное отображение главной страницы",
// 			url:          "/",
// 			setupStorage: func(ms *MockStorage) {
// 				ms.AddGauge("testGauge", 123.45)
// 				ms.AddCounter("testCounter", 100)
// 			},
// 			expectedCode: http.StatusOK,
// 		},
// 		{
// 			name:         "Неверный путь",
// 			url:          "/invalid",
// 			setupStorage: func(ms *MockStorage) {
// 				ms.AddGauge("testGauge", 123.45)
// 			},
// 			expectedCode: http.StatusNotFound,
// 		},
// 		{
// 			name:         "Пустое хранилище",
// 			url:          "/",
// 			setupStorage: func(ms *MockStorage) {},
// 			expectedCode: http.StatusOK,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockStorage := NewMockStorage()
// 			tt.setupStorage(mockStorage)

// 			stor := NewStor(mockStorage)

// 			// Создаем тестовый запрос
// 			req := httptest.NewRequest("GET", tt.url, nil)

// 			// Создаем ResponseRecorder
// 			rr := httptest.NewRecorder()

// 			// Вызываем обработчик
// 			stor.HandleMain(rr, req)

// 			// Проверяем код ответа
// 			if rr.Code != tt.expectedCode {
// 				t.Errorf("HandleMain() код ответа = %v, ожидается %v", rr.Code, tt.expectedCode)
// 			}

// 			// Проверяем Content-Type для успешных ответов
// 			if tt.expectedCode == http.StatusOK {
// 				contentType := rr.Header().Get("Content-Type")
// 				if !strings.Contains(contentType, "text/html") {
// 					t.Errorf("HandleMain() Content-Type = %v, ожидается text/html", contentType)
// 				}
// 			}
// 		})
// 	}
// }

// Дополнительные тесты для проверки граничных случаев
func TestStor_HandleMetricUpdate_EdgeCases(t *testing.T) {
	mockStorage := NewMockStorage()
	stor := NewStor(mockStorage)

	t.Run("Gauge с нулевым значением", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/update/gauge/zeroGauge/0", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("metricType", "gauge")
		rctx.URLParams.Add("metricName", "zeroGauge")
		rctx.URLParams.Add("metricValue", "0")
		*req = *req.WithContext(contextWithRouteContext(req.Context(), rctx))

		rr := httptest.NewRecorder()
		stor.HandleMetricUpdate(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("HandleMetricUpdate() с нулевым значением код ответа = %v, ожидается %v", rr.Code, http.StatusOK)
		}
	})

	t.Run("Counter с отрицательным значением", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/update/counter/negativeCounter/-50", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("metricType", "counter")
		rctx.URLParams.Add("metricName", "negativeCounter")
		rctx.URLParams.Add("metricValue", "-50")
		*req = *req.WithContext(contextWithRouteContext(req.Context(), rctx))

		rr := httptest.NewRecorder()
		stor.HandleMetricUpdate(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("HandleMetricUpdate() с отрицательным значением код ответа = %v, ожидается %v", rr.Code, http.StatusOK)
		}
	})

	t.Run("Gauge с очень большим значением", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/update/bigGauge/1.7976931348623157e+308", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("metricType", "gauge")
		rctx.URLParams.Add("metricName", "bigGauge")
		rctx.URLParams.Add("metricValue", "1.7976931348623157e+308")
		*req = *req.WithContext(contextWithRouteContext(req.Context(), rctx))

		rr := httptest.NewRecorder()
		stor.HandleMetricUpdate(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("HandleMetricUpdate() с большим значением код ответа = %v, ожидается %v", rr.Code, http.StatusOK)
		}
	})
}

func TestStor_HandleMetricGet_EdgeCases(t *testing.T) {
	mockStorage := NewMockStorage()
	stor := NewStor(mockStorage)

	// Добавляем метрики для тестирования
	mockStorage.AddGauge("zeroGauge", 0.0)
	mockStorage.AddCounter("zeroCounter", 0)
	mockStorage.AddGauge("negativeGauge", -123.45)
	mockStorage.AddCounter("negativeCounter", -100)

	tests := []struct {
		name         string
		metricType   string
		metricName   string
		expectedBody string
	}{
		{
			name:         "Gauge с нулевым значением",
			metricType:   "gauge",
			metricName:   "zeroGauge",
			expectedBody: "0",
		},
		{
			name:         "Counter с нулевым значением",
			metricType:   "counter",
			metricName:   "zeroCounter",
			expectedBody: "0",
		},
		{
			name:         "Gauge с отрицательным значением",
			metricType:   "gauge",
			metricName:   "negativeGauge",
			expectedBody: "-123.45",
		},
		{
			name:         "Counter с отрицательным значением",
			metricType:   "counter",
			metricName:   "negativeCounter",
			expectedBody: "-100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/value/"+tt.metricType+"/"+tt.metricName, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("metricType", tt.metricType)
			rctx.URLParams.Add("metricName", tt.metricName)
			*req = *req.WithContext(contextWithRouteContext(req.Context(), rctx))

			rr := httptest.NewRecorder()
			stor.HandleMetricGet(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("HandleMetricGet() код ответа = %v, ожидается %v", rr.Code, http.StatusOK)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("HandleMetricGet() тело ответа = %v, ожидается %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

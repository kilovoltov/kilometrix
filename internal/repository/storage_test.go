package repository

import (
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	
	if storage == nil {
		t.Fatal("NewMemStorage() returned nil")
	}
	
	if storage.metricsGauge == nil {
		t.Error("metricsGauge map is nil")
	}
	
	if storage.metricsCounter == nil {
		t.Error("metricsCounter map is nil")
	}
	
	if len(storage.metricsGauge) != 0 {
		t.Error("metricsGauge should be empty initially")
	}
	
	if len(storage.metricsCounter) != 0 {
		t.Error("metricsCounter should be empty initially")
	}
}

func TestAddGauge(t *testing.T) {
	storage := NewMemStorage()
	
	tests := []struct {
		name  string
		value float64
	}{
		{"test_gauge", 42.5},
		{"negative_gauge", -10.0},
		{"zero_gauge", 0.0},
		{"large_gauge", 1e10},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.AddGauge(tt.name, tt.value)
			if err != nil {
				t.Errorf("AddGauge() returned error: %v", err)
			}
			
			// Проверяем, что метрика была добавлена
			gauge, err := storage.GetGauge(tt.name)
			if err != nil {
				t.Errorf("GetGauge() returned error for added metric: %v", err)
			}
			
			if gauge != tt.value {
				t.Errorf("GetGauge() returned %v, want %v", gauge, tt.value)
			}
		})
	}
}

func TestAddGauge_Overwrite(t *testing.T) {
	storage := NewMemStorage()
	
	// Добавляем метрику первый раз
	err := storage.AddGauge("test", 10.0)
	if err != nil {
		t.Fatal("AddGauge() failed:", err)
	}
	
	// Перезаписываем значение
	err = storage.AddGauge("test", 20.0)
	if err != nil {
		t.Fatal("AddGauge() failed:", err)
	}
	
	// Проверяем, что значение обновилось
	gauge, err := storage.GetGauge("test")
	if err != nil {
		t.Fatal("GetGauge() failed:", err)
	}
	
	if gauge != 20.0 {
		t.Errorf("GetGauge() returned %v, want %v", gauge, 20.0)
	}
}

func TestAddCounter(t *testing.T) {
	storage := NewMemStorage()
	
	tests := []struct {
		name  string
		value int64
	}{
		{"test_counter", 42},
		{"negative_counter", -10},
		{"zero_counter", 0},
		{"large_counter", 1e6},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.AddCounter(tt.name, tt.value)
			if err != nil {
				t.Errorf("AddCounter() returned error: %v", err)
			}
			
			// Проверяем, что метрика была добавлена
			counter, err := storage.GetCounter(tt.name)
			if err != nil {
				t.Errorf("GetCounter() returned error for added metric: %v", err)
			}
			
			if counter != tt.value {
				t.Errorf("GetCounter() returned %v, want %v", counter, tt.value)
			}
		})
	}
}

func TestAddCounter_Accumulate(t *testing.T) {
	storage := NewMemStorage()
	
	// Добавляем значение первый раз
	err := storage.AddCounter("test", 10)
	if err != nil {
		t.Fatal("AddCounter() failed:", err)
	}
	
	// Добавляем значение второй раз
	err = storage.AddCounter("test", 5)
	if err != nil {
		t.Fatal("AddCounter() failed:", err)
	}
	
	// Проверяем, что значения суммируются
	counter, err := storage.GetCounter("test")
	if err != nil {
		t.Fatal("GetCounter() failed:", err)
	}
	
	if counter != 15 {
		t.Errorf("GetCounter() returned %v, want %v", counter, 15)
	}
}

func TestAddCounter_MultipleAdditions(t *testing.T) {
	storage := NewMemStorage()
	
	values := []int64{10, 20, 30, -5, 15}
	expected := int64(70)
	
	for _, value := range values {
		err := storage.AddCounter("test", value)
		if err != nil {
			t.Fatal("AddCounter() failed:", err)
		}
	}
	
	counter, err := storage.GetCounter("test")
	if err != nil {
		t.Fatal("GetCounter() failed:", err)
	}
	
	if counter != expected {
		t.Errorf("GetCounter() returned %v, want %v", counter, expected)
	}
}

func TestGetGauge_NotFound(t *testing.T) {
	storage := NewMemStorage()
	
	_, err := storage.GetGauge("nonexistent")
	if err == nil {
		t.Error("GetGauge() should return error for nonexistent metric")
	}
	
	if err.Error() != "metric nonexistent not found" {
		t.Errorf("GetGauge() error message: %v", err)
	}
}

func TestGetCounter_NotFound(t *testing.T) {
	storage := NewMemStorage()
	
	_, err := storage.GetCounter("nonexistent")
	if err == nil {
		t.Error("GetCounter() should return error for nonexistent metric")
	}
	
	if err.Error() != "metric nonexistent not found" {
		t.Errorf("GetCounter() error message: %v", err)
	}
}

func TestGetGaugesNames(t *testing.T) {
	storage := NewMemStorage()
	
	// Проверяем пустой список
	names := storage.GetGaugesNames()
	if len(names) != 0 {
		t.Errorf("GetGaugesNames() returned %d names, want 0", len(names))
	}
	
	// Добавляем метрики
	storage.AddGauge("gauge1", 10.0)
	storage.AddGauge("gauge2", 20.0)
	storage.AddGauge("gauge3", 30.0)
	
	names = storage.GetGaugesNames()
	if len(names) != 3 {
		t.Errorf("GetGaugesNames() returned %d names, want 3", len(names))
	}
	
	// Проверяем, что все имена присутствуют
	expectedNames := map[string]bool{
		"gauge1": false,
		"gauge2": false,
		"gauge3": false,
	}
	
	for _, name := range names {
		if _, exists := expectedNames[name]; !exists {
			t.Errorf("GetGaugesNames() returned unexpected name: %s", name)
		}
		expectedNames[name] = true
	}
	
	// Проверяем, что все ожидаемые имена были найдены
	for name, found := range expectedNames {
		if !found {
			t.Errorf("GetGaugesNames() missing expected name: %s", name)
		}
	}
}

func TestGetCounterNames(t *testing.T) {
	storage := NewMemStorage()
	
	// Проверяем пустой список
	names := storage.GetCounterNames()
	if len(names) != 0 {
		t.Errorf("GetCounterNames() returned %d names, want 0", len(names))
	}
	
	// Добавляем метрики
	storage.AddCounter("counter1", 10)
	storage.AddCounter("counter2", 20)
	storage.AddCounter("counter3", 30)
	
	names = storage.GetCounterNames()
	if len(names) != 3 {
		t.Errorf("GetCounterNames() returned %d names, want 3", len(names))
	}
	
	// Проверяем, что все имена присутствуют
	expectedNames := map[string]bool{
		"counter1": false,
		"counter2": false,
		"counter3": false,
	}
	
	for _, name := range names {
		if _, exists := expectedNames[name]; !exists {
			t.Errorf("GetCounterNames() returned unexpected name: %s", name)
		}
		expectedNames[name] = true
	}
	
	// Проверяем, что все ожидаемые имена были найдены
	for name, found := range expectedNames {
		if !found {
			t.Errorf("GetCounterNames() missing expected name: %s", name)
		}
	}
}

func TestMixedMetrics(t *testing.T) {
	storage := NewMemStorage()
	
	// Добавляем смешанные метрики
	storage.AddGauge("gauge1", 10.5)
	storage.AddCounter("counter1", 5)
	storage.AddGauge("gauge2", 20.5)
	storage.AddCounter("counter2", 10)
	
	// Проверяем Gauge метрики
	gaugeNames := storage.GetGaugesNames()
	if len(gaugeNames) != 2 {
		t.Errorf("GetGaugesNames() returned %d names, want 2", len(gaugeNames))
	}
	
	// Проверяем Counter метрики
	counterNames := storage.GetCounterNames()
	if len(counterNames) != 2 {
		t.Errorf("GetCounterNames() returned %d names, want 2", len(counterNames))
	}
	
	// Проверяем значения
	gauge1, err := storage.GetGauge("gauge1")
	if err != nil {
		t.Fatal("GetGauge() failed:", err)
	}
	if gauge1 != 10.5 {
		t.Errorf("GetGauge(gauge1) returned %v, want %v", gauge1, 10.5)
	}
	
	counter1, err := storage.GetCounter("counter1")
	if err != nil {
		t.Fatal("GetCounter() failed:", err)
	}
	if counter1 != 5 {
		t.Errorf("GetCounter(counter1) returned %v, want %v", counter1, 5)
	}
}

func TestEdgeCases(t *testing.T) {
	storage := NewMemStorage()
	
	// Тест с очень большими числами
	err := storage.AddGauge("large_gauge", 1.7976931348623157e+308)
	if err != nil {
		t.Error("AddGauge() failed for large number:", err)
	}
	
	err = storage.AddCounter("large_counter", 9223372036854775807) // max int64
	if err != nil {
		t.Error("AddCounter() failed for max int64:", err)
	}
	
	// Тест с очень маленькими числами
	err = storage.AddGauge("small_gauge", -1.7976931348623157e+308)
	if err != nil {
		t.Error("AddGauge() failed for small number:", err)
	}
	
	err = storage.AddCounter("small_counter", -9223372036854775808) // min int64
	if err != nil {
		t.Error("AddCounter() failed for min int64:", err)
	}
	
	// Проверяем, что все значения сохранились корректно
	largeGauge, err := storage.GetGauge("large_gauge")
	if err != nil {
		t.Fatal("GetGauge() failed for large_gauge:", err)
	}
	
	largeCounter, err := storage.GetCounter("large_counter")
	if err != nil {
		t.Fatal("GetCounter() failed for large_counter:", err)
	}
	
	smallGauge, err := storage.GetGauge("small_gauge")
	if err != nil {
		t.Fatal("GetGauge() failed for small_gauge:", err)
	}
	
	smallCounter, err := storage.GetCounter("small_counter")
	if err != nil {
		t.Fatal("GetCounter() failed for small_counter:", err)
	}
	
	// Проверяем значения (с учетом точности float64)
	if largeGauge <= 0 {
		t.Errorf("GetGauge(large_gauge) returned %v, want positive large number", largeGauge)
	}
	
	if largeCounter != 9223372036854775807 {
		t.Errorf("GetCounter(large_counter) returned %v, want %v", largeCounter, 9223372036854775807)
	}
	
	if smallGauge >= 0 {
		t.Errorf("GetGauge(small_gauge) returned %v, want negative large number", smallGauge)
	}
	
	if smallCounter != -9223372036854775808 {
		t.Errorf("GetCounter(small_counter) returned %v, want %v", smallCounter, -9223372036854775808)
	}
}
package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type FileStorage struct {
	mem *MemStorage

	filepath string
	interval time.Duration
}

func NewFileStorage(filepath string, interval time.Duration) *FileStorage {
	fs := &FileStorage{
		mem:      NewMemStorage(),
		filepath: filepath,
		interval: interval,
	}

	if interval > 0 {
		go fs.runPeriodicSaver()
	}

	return fs
}

func (f *FileStorage) runPeriodicSaver() {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	done := make(chan bool)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			_ = f.saveToFile() // игнорируем ошибку или логируем
		}
	}
}

func (f *FileStorage) saveToFile() error {
	// Делаем снимок данных
	snapshot := f.Snapshot()
	// snapshot := struct {
	// 	Gauges   map[string]float64 `json:"gauges"`
	// 	Counters map[string]int64   `json:"counters"`
	// }{
	// 	Gauges:   make(map[string]float64, len(f.mem.metricsGauge)),
	// 	Counters: make(map[string]int64, len(f.mem.metricsCounter)),
	// }

	// Записываем в файл
	file, err := os.OpenFile(f.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", f.filepath, err)
	}
	defer file.Close()

	_, err = file.Write(snapshot)
	return err

	// encoder := json.NewEncoder(file)
	// encoder.SetIndent("", "  ")
	// return encoder.Encode(snapshot)
}

func (f *FileStorage) AddGauge(name string, value float64) error {
	if err := f.mem.AddGauge(name, value); err != nil {
		return err
	}

	// Синхронная запись, если interval == 0
	if f.interval == 0 {
		return f.saveToFile()
	}
	return nil
}

func (f *FileStorage) AddCounter(name string, value int64) error {
	if err := f.mem.AddCounter(name, value); err != nil {
		return err
	}

	if f.interval == 0 {
		return f.saveToFile()
	}
	return nil
}

func (f *FileStorage) GetGauge(name string) (float64, error) {
	return f.mem.GetGauge(name)
}

func (f *FileStorage) GetCounter(name string) (int64, error) {
	return f.mem.GetCounter(name)
}

func (f *FileStorage) GetGaugesNames() []string {
	return f.mem.GetGaugesNames()
}

func (f *FileStorage) GetCounterNames() []string {
	return f.mem.GetCounterNames()
}

func (f *FileStorage) Snapshot() []byte {
	return f.mem.Snapshot()
}

// Опционально: метод загрузки из файла при старте
func (f *FileStorage) LoadFromFile() error {
	file, err := os.Open(f.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // файл ещё не существует — нормально
		}
		return err
	}
	defer file.Close()

	var data struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	f.mem.metricsGauge = data.Gauges
	f.mem.metricsCounter = data.Counters

	return nil
}

package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kilovoltov/kilometrix/internal/models"
)

type FileStorage struct {
	mem *MemStorage

	filepath string
	interval time.Duration
	done     chan struct{}
}

func NewFileStorage(filepath string, interval time.Duration) *FileStorage {
	fs := &FileStorage{
		mem:      NewMemStorage(),
		filepath: filepath,
		interval: interval,
		done:     make(chan struct{}),
	}

	if interval > 0 {
		go fs.runPeriodicSaver()
	}

	return fs
}

func (f *FileStorage) runPeriodicSaver() {
	ticker := time.NewTicker(f.interval)
	defer ticker.Stop()

	for {
		select {
		case <-f.done:
			return
		case <-ticker.C:
			err := f.saveToFile()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Close закрывает канал done, что приводит к завершению горутины
func (f *FileStorage) Close() {
	close(f.done)
}

func (f *FileStorage) saveToFile() error {
	// Делаем снимок данных
	snapshot := f.Snapshot()

	// Записываем в файл
	file, err := os.OpenFile(f.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", f.filepath, err)
	}
	defer file.Close()

	// Пишем открывающую скобку
	if _, err := file.WriteString("[\n"); err != nil {
		return err
	}

	for i, m := range snapshot {
		// Сериализуем структуру в JSON без отступов (compact)
		data, err := json.Marshal(m)
		if err != nil {
			return err
		}

		// Пишем строку
		if _, err := file.Write(data); err != nil {
			return err
		}

		// Добавляем запятую, если это НЕ последний элемент
		if i < len(snapshot)-1 {
			if _, err := file.WriteString(",\n"); err != nil {
				return err
			}
		} else {
			// Последний элемент — только перенос строки
			if _, err := file.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	// Пишем закрывающую скобку
	if _, err := file.WriteString("]"); err != nil {
		return err
	}

	fileExistsAndNotEmpty(f.filepath)

	// Запись в формате JSONL
	// encoder := json.NewEncoder(file)
	// for _, m := range snapshot {
	// 	if err := encoder.Encode(m); err != nil {
	// 		return err
	// 	}
	// }

	return nil
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

func (f *FileStorage) Snapshot() []models.Metrics {
	return f.mem.Snapshot()
}

// LoadFromFile (опционально) метод загрузки из файла при старте
func (f *FileStorage) LoadFromFile() error {
	fmt.Printf("+++++++++++++++++ Loading from file: %s\n", f.filepath)
	file, err := os.Open(f.filepath)
	if err != nil {
		fmt.Printf("+++++++++ Something goes wrong: %v", err)
		if os.IsNotExist(err) {
			return nil // если файла нет - это может быть нормаьлно
		}
		return err
	}
	defer file.Close()

	var metrics []models.Metrics
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metrics); err != nil {
		fmt.Println(err)
	}

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			f.AddGauge(m.ID, *m.Value)
		case "counter":
			f.AddCounter(m.ID, *m.Delta)
		}
	}

	return nil
}

func fileExistsAndNotEmpty(filename string) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		fmt.Printf("=============: File doesn't exist: %v", err) // файл не существует
	}
	if err != nil {
		fmt.Printf("=============: Other ERROR: %v", err) // другая ошибка (например, нет прав)
	}

	// Проверяем, что файл не пустой
	if info.Size() > 0 {
		fmt.Printf("=============: File exists and not empty\n")
	}
}

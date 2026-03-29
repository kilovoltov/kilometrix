package sender

import (
	"fmt"
	"sync"

	"github.com/kilovoltov/kilometrix/internal/models"
)

// WorkerPool представляет пул воркеров для отправки метрик
type WorkerPool struct {
	workers   int           // Количество воркеров
	tasksChan chan Task     // Канал для задач
	wg        sync.WaitGroup // WaitGroup для ожидания завершения воркеров
}

// NewWorkerPool создаёт новый пул воркеров
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers:   workers,
		tasksChan: make(chan Task, workers*2), // Буфер в 2 раза больше количества воркеров
	}
}

// Start запускает пул воркеров
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker обрабатывает задачи из канала
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for task := range wp.tasksChan {
		if err := wp.processTask(task); err != nil {
			fmt.Printf("Worker %d: error sending metrics: %v\n", id, err)
		}
	}
}

// processTask обрабатывает одну задачу отправки метрик
func (wp *WorkerPool) processTask(task Task) error {
	// Создаём копию метрик для отправки
	metricsCopy := make(models.Storage, len(task.Metrics))
	for k, v := range task.Metrics {
		metricCopy := *v
		metricsCopy[k] = &metricCopy
	}

	return SendMetricsJSON(task.Server, task.Client, metricsCopy, task.Secret)
}

// Submit отправляет задачу в пул воркеров
// Блокируется, если канал задач полон
func (wp *WorkerPool) Submit(task Task) {
	wp.tasksChan <- task
}

// Stop останавливает пул воркеров
// Закрывает канал задач и ожидает завершения всех воркеров
func (wp *WorkerPool) Stop() {
	close(wp.tasksChan)
	wp.wg.Wait()
}

// WorkersCount возвращает количество воркеров в пуле
func (wp *WorkerPool) WorkersCount() int {
	return wp.workers
}

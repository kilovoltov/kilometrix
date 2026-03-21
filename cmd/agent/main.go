package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/kilovoltov/kilometrix/internal/agent/collector"
	"github.com/kilovoltov/kilometrix/internal/agent/sender"
	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/go-resty/resty/v2"
)

func main() {
	parseFlags()

	// Инициализация хранилища метрик
	metrics := models.Storage{
		"PollCount":       {Name: "PollCount", Type: models.CounterType, Value: ""},
		"Alloc":           {Name: "Alloc", Type: models.GaugeType, Value: ""},
		"BuckHashSys":     {Name: "BuckHashSys", Type: models.GaugeType, Value: ""},
		"Frees":           {Name: "Frees", Type: models.GaugeType, Value: ""},
		"GCCPUFraction":   {Name: "GCCPUFraction", Type: models.GaugeType, Value: ""},
		"GCSys":           {Name: "GCSys", Type: models.GaugeType, Value: ""},
		"HeapAlloc":       {Name: "HeapAlloc", Type: models.GaugeType, Value: ""},
		"HeapIdle":        {Name: "HeapIdle", Type: models.GaugeType, Value: ""},
		"HeapInuse":       {Name: "HeapInuse", Type: models.GaugeType, Value: ""},
		"HeapObjects":     {Name: "HeapObjects", Type: models.GaugeType, Value: ""},
		"HeapReleased":    {Name: "HeapReleased", Type: models.GaugeType, Value: ""},
		"HeapSys":         {Name: "HeapSys", Type: models.GaugeType, Value: ""},
		"LastGC":          {Name: "LastGC", Type: models.GaugeType, Value: ""},
		"Lookups":         {Name: "Lookups", Type: models.GaugeType, Value: ""},
		"MCacheInuse":     {Name: "MCacheInuse", Type: models.GaugeType, Value: ""},
		"MCacheSys":       {Name: "MCacheSys", Type: models.GaugeType, Value: ""},
		"MSpanInuse":      {Name: "MSpanInuse", Type: models.GaugeType, Value: ""},
		"MSpanSys":        {Name: "MSpanSys", Type: models.GaugeType, Value: ""},
		"Mallocs":         {Name: "Mallocs", Type: models.GaugeType, Value: ""},
		"NextGC":          {Name: "NextGC", Type: models.GaugeType, Value: ""},
		"NumForcedGC":     {Name: "NumForcedGC", Type: models.GaugeType, Value: ""},
		"NumGC":           {Name: "NumGC", Type: models.GaugeType, Value: ""},
		"OtherSys":        {Name: "OtherSys", Type: models.GaugeType, Value: ""},
		"PauseTotalNs":    {Name: "PauseTotalNs", Type: models.GaugeType, Value: ""},
		"StackInuse":      {Name: "StackInuse", Type: models.GaugeType, Value: ""},
		"StackSys":        {Name: "StackSys", Type: models.GaugeType, Value: ""},
		"Sys":             {Name: "Sys", Type: models.GaugeType, Value: ""},
		"TotalAlloc":      {Name: "TotalAlloc", Type: models.GaugeType, Value: ""},
		"RandomValue":     {Name: "RandomValue", Type: models.GaugeType, Value: ""},
		"TotalMemory":     {Name: "TotalMemory", Type: models.GaugeType, Value: ""},
		"FreeMemory":      {Name: "FreeMemory", Type: models.GaugeType, Value: ""},
		"CPUutilization1": {Name: "CPUutilization1", Type: models.GaugeType, Value: ""},
	}

	// Создаём контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// HTTP клиент
	client := resty.New().
		SetTimeout(2 * time.Second)

	// Создаём worker pool
	workerPool := sender.NewWorkerPool(rateLimit)
	workerPool.Start()
	defer workerPool.Stop()

	// WaitGroup для ожидания завершения горутин
	var wg sync.WaitGroup

	// Запускаем коллектор runtime метрик в отдельной горутине
	wg.Add(1)
	go runCollector(ctx, &wg, metrics)

	// Запускаем коллектор системных метрик в отдельной горутине
	wg.Add(1)
	go runSystemCollector(ctx, &wg, metrics)

	// Запускаем диспетчер в отдельной горутине
	wg.Add(1)
	go runDispatcher(ctx, &wg, metrics, workerPool, client)

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал завершения
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping agent...")

	// Отменяем контекст для остановки горутин
	cancel()

	// Ожидаем завершения всех горутин
	wg.Wait()
	fmt.Println("Agent stopped gracefully")
}

// runCollector собирает runtime метрики
func runCollector(ctx context.Context, wg *sync.WaitGroup, metrics models.Storage) {
	defer wg.Done()

	var counter int64
	tickerPoll := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer tickerPoll.Stop()

	for {
		select {
		case <-ctx.Done():
			// Контекст отменён, завершаем работу
			return
		case <-tickerPoll.C:
			// Сбор runtime метрик
			counter += 1
			metrics["PollCount"].Value = strconv.FormatInt(counter, 10)
			collector.CollectRuntimeMetrics(metrics)
		}
	}
}

// runSystemCollector собирает системные метрики
func runSystemCollector(ctx context.Context, wg *sync.WaitGroup, metrics models.Storage) {
	defer wg.Done()

	tickerPoll := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer tickerPoll.Stop()

	for {
		select {
		case <-ctx.Done():
			// Контекст отменён, завершаем работу
			return
		case <-tickerPoll.C:
			// Сбор системных метрик
			if err := collector.CollectSystemMetrics(metrics); err != nil {
				fmt.Printf("Error collecting system metrics: %v\n", err)
			}
		}
	}
}

// runDispatcher отправляет копию всех метрик в worker pool
func runDispatcher(ctx context.Context, wg *sync.WaitGroup, metrics models.Storage, workerPool *sender.WorkerPool, client *resty.Client) {
	defer wg.Done()

	tickerReport := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer tickerReport.Stop()

	for {
		select {
		case <-ctx.Done():
			// Контекст отменён, завершаем работу
			return
		case <-tickerReport.C:
			// Создаём копию всех метрик для отправки
			metricsCopy := make(models.Storage, len(metrics))
			for k, v := range metrics {
				metricCopy := *v
				metricsCopy[k] = &metricCopy
			}

			// Создаём задачу для воркера
			task := sender.Task{
				Metrics: metricsCopy,
				Server:  addr,
				Client:  client,
				Secret:  secretKey,
			}
			// Отправляем задачу в worker pool
			workerPool.Submit(task)
		}
	}
}

package main

import (
    "flag"
    "fmt"
    "log"
    "strconv"
    "time"

    "github.com/kilovoltov/kilometrix/internal/agent/collector"
    "github.com/kilovoltov/kilometrix/internal/agent/sender"
    "github.com/kilovoltov/kilometrix/internal/models"

    "github.com/go-resty/resty/v2"
    "github.com/caarlos0/env/v6"
)

type Config struct {
    Address         string `env:"ADDRESS"`
    PollInterval   int64  `env:"POLL_INTERVAL"`
    ReportInterval int    `env:"REPORT_INTERVAL"`
}

func main() {
    // параметры запуска: адрес сервера, интервалы сбора и отправки метрик
    var addr string
    var pollInterval int64
    var reportInterval int
    flag.StringVar(&addr, "a", "localhost:8080", "address of the server")
    flag.Int64Var(&pollInterval, "p", 2, "poll interval, sec")
    flag.IntVar(&reportInterval, "r", 10, "report interval, sec")
    flag.Parse()

    var cfg Config
    err := env.Parse(&cfg)
    if err != nil {
        log.Fatal(err)
    }

    // если есть переменные окружения, то они перезаписывают значения флагов
    if cfg.Address != "" {
        addr = cfg.Address
    }
    if cfg.PollInterval != 0 {
        pollInterval = cfg.PollInterval
    }
    if cfg.ReportInterval != 0 {
        reportInterval = cfg.ReportInterval
    }

    metrics := models.Storage{
        "PollCount":     {Name: "PollCount", Type: models.CounterType, Value: ""},
        "Alloc":         {Name: "Alloc", Type: models.GaugeType, Value: ""},
        "BuckHashSys":   {Name: "BuckHashSys", Type: models.GaugeType, Value: ""},
        "Frees":         {Name: "Frees", Type: models.CounterType, Value: ""},
        "GCCPUFraction": {Name: "GCCPUFraction", Type: models.GaugeType, Value: ""},
        "GCSys":         {Name: "GCSys", Type: models.GaugeType, Value: ""},
        "HeapAlloc":     {Name: "HeapAlloc", Type: models.GaugeType, Value: ""},
        "HeapIdle":      {Name: "HeapIdle", Type: models.GaugeType, Value: ""},
        "HeapInuse":     {Name: "HeapInuse", Type: models.GaugeType, Value: ""},
        "HeapObjects":   {Name: "HeapObjects", Type: models.GaugeType, Value: ""},
        "HeapReleased":  {Name: "HeapReleased", Type: models.GaugeType, Value: ""},
        "HeapSys":       {Name: "HeapSys", Type: models.GaugeType, Value: ""},
        "LastGC":        {Name: "LastGC", Type: models.GaugeType, Value: ""},
        "Lookups":       {Name: "Lookups", Type: models.CounterType, Value: ""},
        "MCacheInuse":   {Name: "MCacheInuse", Type: models.GaugeType, Value: ""},
        "MCacheSys":     {Name: "MCacheSys", Type: models.GaugeType, Value: ""},
        "MSpanInuse":    {Name: "MSpanInuse", Type: models.GaugeType, Value: ""},
        "MSpanSys":      {Name: "MSpanSys", Type: models.GaugeType, Value: ""},
        "Mallocs":       {Name: "Mallocs", Type: models.GaugeType, Value: ""},
        "NextGC":        {Name: "NextGC", Type: models.GaugeType, Value: ""},
        "NumForcedGC":   {Name: "NumForcedGC", Type: models.GaugeType, Value: ""},
        "NumGC":         {Name: "NumGC", Type: models.GaugeType, Value: ""},
        "OtherSys":      {Name: "OtherSys", Type: models.GaugeType, Value: ""},
        "PauseTotalNs":  {Name: "PauseTotalNs", Type: models.GaugeType, Value: ""},
        "StackInuse":    {Name: "StackInuse", Type: models.GaugeType, Value: ""},
        "StackSys":      {Name: "StackSys", Type: models.GaugeType, Value: ""},
        "Sys":           {Name: "Sys", Type: models.GaugeType, Value: ""},
        "TotalAlloc":    {Name: "TotalAlloc", Type: models.GaugeType, Value: ""},
        "RandomValue":   {Name: "RandomValue", Type: models.GaugeType, Value: ""},
    }

    var counter int64

    client := resty.New().
        SetTimeout(2 * time.Second)

    tickerPoll := time.NewTicker(time.Duration(pollInterval) * time.Second)
    tickerReport := time.NewTicker(time.Duration(reportInterval) * time.Second)
    defer tickerPoll.Stop()
    defer tickerReport.Stop()

    for {
        select {
        case <-tickerPoll.C:
            // Сбор метрик
            counter += 1
            metrics["PollCount"].Value = strconv.FormatInt(counter, 10)
            collector.CollectRuntimeMetrics(metrics)

        // Отправка метрик
        case <-tickerReport.C:
            for _, metric := range metrics {
                if err := sender.SendMetric(addr, client, *metric); err != nil {
                    fmt.Printf("Error sending metric %s: %v\n", metric.Name, err)
                }
            }
            counter = 0
        }
    }
}

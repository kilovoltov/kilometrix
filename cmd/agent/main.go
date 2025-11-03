package main

import (
    "flag"
    "fmt"
    "math/rand/v2"
    "runtime"
    "strconv"
    "time"

    "github.com/go-resty/resty/v2"
)

type MetricType string

const (
    GaugeType   MetricType = "gauge"
    CounterType MetricType = "counter"
)

// Metric — описание метрики
type Metric struct {
    Name  string
    Type  MetricType
    Value string
}

type Storage map[string]*Metric

// Функция сбора метрик из runtime.MemStats
func CollectRuntimeMetrics(stor Storage) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    stor["Alloc"].Value = strconv.FormatUint(memStats.Alloc, 10)
    stor["BuckHashSys"].Value = strconv.FormatUint(memStats.BuckHashSys, 10)
    stor["Frees"].Value = strconv.FormatUint(memStats.Frees, 10)
    stor["GCCPUFraction"].Value = strconv.FormatFloat(memStats.GCCPUFraction, 'f', -1, 64)
    stor["GCSys"].Value = strconv.FormatUint(memStats.GCSys, 10)
    stor["HeapAlloc"].Value = strconv.FormatUint(memStats.HeapAlloc, 10)
    stor["HeapIdle"].Value = strconv.FormatUint(memStats.HeapIdle, 10)
    stor["HeapInuse"].Value = strconv.FormatUint(memStats.HeapInuse, 10)
    stor["HeapObjects"].Value = strconv.FormatUint(memStats.HeapObjects, 10)
    stor["HeapReleased"].Value = strconv.FormatUint(memStats.HeapReleased, 10)
    stor["HeapSys"].Value = strconv.FormatUint(memStats.HeapSys, 10)
    stor["LastGC"].Value = strconv.FormatUint(memStats.LastGC, 10)
    stor["Lookups"].Value = strconv.FormatUint(memStats.Lookups, 10)
    stor["MCacheInuse"].Value = strconv.FormatUint(memStats.MCacheInuse, 10)
    stor["MCacheSys"].Value = strconv.FormatUint(memStats.MCacheSys, 10)
    stor["MSpanInuse"].Value = strconv.FormatUint(memStats.MSpanInuse, 10)
    stor["MSpanSys"].Value = strconv.FormatUint(memStats.MSpanSys, 10)
    stor["Mallocs"].Value = strconv.FormatUint(memStats.Mallocs, 10)
    stor["NextGC"].Value = strconv.FormatUint(memStats.NextGC, 10)
    stor["NumForcedGC"].Value = strconv.FormatUint(uint64(memStats.NumForcedGC), 10)
    stor["NumGC"].Value = strconv.FormatUint(uint64(memStats.NumGC), 10)
    stor["OtherSys"].Value = strconv.FormatUint(memStats.OtherSys, 10)
    stor["PauseTotalNs"].Value = strconv.FormatUint(memStats.PauseTotalNs, 10)
    stor["StackInuse"].Value = strconv.FormatUint(memStats.StackInuse, 10)
    stor["StackSys"].Value = strconv.FormatUint(memStats.StackSys, 10)
    stor["Sys"].Value = strconv.FormatUint(memStats.Sys, 10)
    stor["TotalAlloc"].Value = strconv.FormatUint(memStats.TotalAlloc, 10)
    stor["RandomValue"].Value = strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
}

// отправляет одну метрику на сервер
func SendMetric(serverAddress string, client *resty.Client, metric Metric) error {
    url := fmt.Sprintf("http://%s/update/%s/%s/%s",
        serverAddress,
        metric.Type,
        metric.Name,
        metric.Value,
    )

    resp, err := client.R().
        SetHeader("Content-Type", "text/plain").
        Post(url)

    if err != nil {
        return fmt.Errorf("failed to send metric %s: %w", metric.Name, err)
    }
    if resp.StatusCode() != 200 {
        return fmt.Errorf("non-200 status for metric %s: %s", metric.Name, resp.Status())
    }
    return nil
}

func main() {
    // переменная для адреса сервера со значением по умолчанию
    var addr = flag.String("a", "localhost:8080", "address of the server")
    var pollInterval = flag.Int64("p", 2, "poll interval, sec")
    var reportInterval = flag.Int("r", 10, "report interval, sec")
    flag.Parse()

    metrics := Storage{
        "PollCount": {"PollCount", CounterType, ""},
        "Alloc": {"Alloc", GaugeType, ""},
        "BuckHashSys": {"BuckHashSys", GaugeType, ""},
        "Frees": {"Frees", CounterType, ""},
        "GCCPUFraction": {"GCCPUFraction", GaugeType, ""},
        "GCSys": {"GCSys", GaugeType, ""},
        "HeapAlloc": {"HeapAlloc", GaugeType, ""},
        "HeapIdle": {"HeapIdle", GaugeType, ""},
        "HeapInuse": {"HeapInuse", GaugeType, ""},
        "HeapObjects": {"HeapObjects", GaugeType, ""},
        "HeapReleased": {"HeapReleased", GaugeType, ""},
        "HeapSys": {"HeapSys", GaugeType, ""},
        "LastGC": {"LastGC", GaugeType, ""},
        "Lookups": {"Lookups", CounterType, ""},
        "MCacheInuse": {"MCacheInuse", GaugeType, ""},
        "MCacheSys": {"MCacheSys", GaugeType, ""},
        "MSpanInuse": {"MSpanInuse", GaugeType, ""},
        "MSpanSys": {"MSpanSys", GaugeType, ""},
        "Mallocs": {"Mallocs", GaugeType, ""},
        "NextGC": {"NextGC", GaugeType, ""},
        "NumForcedGC": {"NumForcedGC", GaugeType, ""},
        "NumGC": {"NumGC", GaugeType, ""},
        "OtherSys": {"OtherSys", GaugeType, ""},
        "PauseTotalNs": {"PauseTotalNs", GaugeType, ""},
        "StackInuse": {"StackInuse", GaugeType, ""},
        "StackSys": {"StackSys", GaugeType, ""},
        "Sys": {"Sys", GaugeType, ""},
        "TotalAlloc": {"TotalAlloc", GaugeType, ""},
        "RandomValue": {"RandomValue", GaugeType, ""},
    }

    var counter int64

    client := resty.New().
        SetTimeout(2 * time.Second)

    tickerPoll := time.NewTicker(time.Duration(*pollInterval) * time.Second)
    tickerReport := time.NewTicker(time.Duration(*reportInterval) * time.Second)
    defer tickerPoll.Stop()
    defer tickerReport.Stop()

    // var i int64
    for {
        // fmt.Println(i)
        // if i%*pollInterval == 0 {
        select {
        case <- tickerPoll.C:
            // Сбор метрик
            counter += 1
            // countSum, _ := strconv.ParseInt(metrics["PollCount"].Value, 10, 64)
            metrics["PollCount"].Value = strconv.FormatInt(counter, 10)
            CollectRuntimeMetrics(metrics)
            fmt.Println("Собрал метрики")
        

        // Отправка метрик
        // if i%*reportInterval == 0 {
        case <- tickerReport.C:
            // fmt.Printf("Metrics %v\n", metrics)
            for _, metric := range metrics {
                if err := SendMetric(*addr, client, *metric); err != nil {
                    fmt.Printf("Error sending metric %s: %v\n", metric.Name, err)
                }
            }
            counter = 0
        }
        // time.Sleep(time.Second)
    }
}

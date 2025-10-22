package main

import (
    "fmt"
    "math/rand"
    "runtime"
    "strconv"
    "time"

    "github.com/go-resty/resty/v2"
    flag "github.com/spf13/pflag"
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

type Storage [29]Metric

// Функция сбора метрик из runtime.MemStats
func CollectRuntimeMetrics(s *Storage) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    s[1].Value = strconv.FormatUint(memStats.Alloc, 10)
    s[2].Value = strconv.FormatUint(memStats.BuckHashSys, 10)
    s[3].Value = strconv.FormatUint(memStats.Frees, 10)
    s[4].Value = strconv.FormatFloat(memStats.GCCPUFraction, 'f', -1, 64)
    s[5].Value = strconv.FormatUint(memStats.GCSys, 10)
    s[6].Value = strconv.FormatUint(memStats.HeapAlloc, 10)
    s[7].Value = strconv.FormatUint(memStats.HeapIdle, 10)
    s[8].Value = strconv.FormatUint(memStats.HeapInuse, 10)
    s[9].Value = strconv.FormatUint(memStats.HeapObjects, 10)
    s[10].Value = strconv.FormatUint(memStats.HeapReleased, 10)
    s[11].Value = strconv.FormatUint(memStats.HeapSys, 10)
    s[12].Value = strconv.FormatUint(memStats.LastGC, 10)
    s[13].Value = strconv.FormatUint(memStats.Lookups, 10)
    s[14].Value = strconv.FormatUint(memStats.MCacheInuse, 10)
    s[15].Value = strconv.FormatUint(memStats.MCacheSys, 10)
    s[16].Value = strconv.FormatUint(memStats.MSpanInuse, 10)
    s[17].Value = strconv.FormatUint(memStats.MSpanSys, 10)
    s[18].Value = strconv.FormatUint(memStats.Mallocs, 10)
    s[19].Value = strconv.FormatUint(memStats.NextGC, 10)
    s[20].Value = strconv.FormatUint(uint64(memStats.NumForcedGC), 10)
    s[21].Value = strconv.FormatUint(uint64(memStats.NumGC), 10)
    s[22].Value = strconv.FormatUint(memStats.OtherSys, 10)
    s[23].Value = strconv.FormatUint(memStats.PauseTotalNs, 10)
    s[24].Value = strconv.FormatUint(memStats.StackInuse, 10)
    s[25].Value = strconv.FormatUint(memStats.StackSys, 10)
    s[26].Value = strconv.FormatUint(memStats.Sys, 10)
    s[27].Value = strconv.FormatUint(memStats.TotalAlloc, 10)
    s[28].Value = strconv.FormatFloat(rand.Float64(), 'f', -1, 64)
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
    var reportInterval = flag.Int64("r", 10, "report interval, sec")
    flag.Parse()

    metrics := Storage{
        {"PollCount", CounterType, ""},
        {"Alloc", GaugeType, ""},
        {"BuckHashSys", GaugeType, ""},
        {"Frees", CounterType, ""},
        {"GCCPUFraction", GaugeType, ""},
        {"GCSys", GaugeType, ""},
        {"HeapAlloc", GaugeType, ""},
        {"HeapIdle", GaugeType, ""},
        {"HeapInuse", GaugeType, ""},
        {"HeapObjects", GaugeType, ""},
        {"HeapReleased", GaugeType, ""},
        {"HeapSys", GaugeType, ""},
        {"LastGC", GaugeType, ""},
        {"Lookups", CounterType, ""},
        {"MCacheInuse", GaugeType, ""},
        {"MCacheSys", GaugeType, ""},
        {"MSpanInuse", GaugeType, ""},
        {"MSpanSys", GaugeType, ""},
        {"Mallocs", GaugeType, ""},
        {"NextGC", GaugeType, ""},
        {"NumForcedGC", GaugeType, ""},
        {"NumGC", GaugeType, ""},
        {"OtherSys", GaugeType, ""},
        {"PauseTotalNs", GaugeType, ""},
        {"StackInuse", GaugeType, ""},
        {"StackSys", GaugeType, ""},
        {"Sys", GaugeType, ""},
        {"TotalAlloc", GaugeType, ""},
        {"RandomValue", GaugeType, ""},
    }

    var counter int64

    client := resty.New().
        SetTimeout(2 * time.Second)

    var i int64

    for ; ; i++ {
        // fmt.Println(i)
        if i%*pollInterval == 0 {
            // Сбор метрик
            counter += 1
            countSum, _ := strconv.ParseInt(metrics[0].Value, 10, 64)
            metrics[0].Value = strconv.FormatInt(countSum+counter, 10)
            CollectRuntimeMetrics(&metrics)
            // fmt.Println("Собрал метрики")
        }

        // Отправка метрик
        if i%*reportInterval == 0 {
            // fmt.Printf("Metrics %v\n", metrics)
            for _, metric := range metrics {
                if err := SendMetric(*addr, client, metric); err != nil {
                    fmt.Printf("Error sending metric %s: %v\n", metric.Name, err)
                }
            }
        }
        time.Sleep(time.Second)
    }
}

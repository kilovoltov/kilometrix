package collector

import (
    "math/rand/v2"
    "runtime"
    "strconv"
	"github.com/kilovoltov/kilometrix/internal/models"
)

// Функция сбора метрик из runtime.MemStats
func CollectRuntimeMetrics(stor models.Storage) {
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
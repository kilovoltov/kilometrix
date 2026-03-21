package collector

import (
	"fmt"
	"time"

	"github.com/kilovoltov/kilometrix/internal/models"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// CollectSystemMetrics собирает системные метрики с помощью gopsutil
func CollectSystemMetrics(stor models.Storage) error {
	// Сбор метрик памяти
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get virtual memory stats: %w", err)
	}

	stor["TotalMemory"].Value = fmt.Sprintf("%d", vmStat.Total)
	stor["FreeMemory"].Value = fmt.Sprintf("%d", vmStat.Free)

	// Сбор метрик CPU (утилизация за 1 секунду)
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("failed to get cpu percent: %w", err)
	}

	if len(percent) > 0 {
		stor["CPUutilization1"].Value = fmt.Sprintf("%.2f", percent[0])
	}

	return nil
}

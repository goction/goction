package viewmodels

import (
	"goction/internal/config"
	"goction/internal/stats"
	"time"
)

type SystemStats struct {
	CPUUsage    float64
	MemoryUsage float64
	Uptime      time.Duration
}

type DashboardData struct {
	Config         *config.Config
	Stats          map[string]*stats.GoctionStats
	History        map[string][]stats.ExecutionRecord
	RecentLogs     []string
	GoctionVersion string
}

func Reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

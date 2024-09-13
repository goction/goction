package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type GoctionStats struct {
	TotalCalls      int           `json:"total_calls"`
	SuccessfulCalls int           `json:"successful_calls"`
	TotalDuration   time.Duration `json:"total_duration"`
	LastExecuted    time.Time     `json:"last_executed"`
}

type ExecutionRecord struct {
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Result    string        `json:"result"`
}

type Manager struct {
	statsFile string
	stats     map[string]*GoctionStats
	history   map[string][]ExecutionRecord
	mu        sync.RWMutex
}

func NewManager(statsFile string) (*Manager, error) {
	m := &Manager{
		statsFile: statsFile,
		stats:     make(map[string]*GoctionStats),
		history:   make(map[string][]ExecutionRecord),
	}

	if err := m.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load stats: %w", err)
	}

	return m, nil
}

func (m *Manager) RecordExecution(name string, duration time.Duration, success bool, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stats == nil {
		m.stats = make(map[string]*GoctionStats)
	}
	if m.history == nil {
		m.history = make(map[string][]ExecutionRecord)
	}

	stats, ok := m.stats[name]
	if !ok {
		stats = &GoctionStats{}
		m.stats[name] = stats
	}

	stats.TotalCalls++
	if success {
		stats.SuccessfulCalls++
	}
	stats.TotalDuration += duration
	stats.LastExecuted = time.Now()

	record := ExecutionRecord{
		Timestamp: time.Now(),
		Duration:  duration,
		Status:    "success",
		Result:    result,
	}
	if !success {
		record.Status = "failure"
	}

	m.history[name] = append(m.history[name], record)

	if err := m.save(); err != nil {
		fmt.Printf("Failed to save stats: %v\n", err)
	}
}

func (m *Manager) GetStats(name string) (*GoctionStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, ok := m.stats[name]
	return stats, ok
}

func (m *Manager) GetAllStats() map[string]*GoctionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allStats := make(map[string]*GoctionStats)
	for name, stats := range m.stats {
		allStats[name] = &GoctionStats{
			TotalCalls:      stats.TotalCalls,
			SuccessfulCalls: stats.SuccessfulCalls,
			TotalDuration:   stats.TotalDuration,
			LastExecuted:    stats.LastExecuted,
		}
	}

	return allStats
}

func (m *Manager) GetExecutionHistory(name string) []ExecutionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.history[name]
}

func (m *Manager) load() error {
	file, err := os.Open(m.statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, just return without error
			return nil
		}
		return fmt.Errorf("failed to open stats file: %w", err)
	}
	defer file.Close()

	var data struct {
		Stats   map[string]*GoctionStats     `json:"stats"`
		History map[string][]ExecutionRecord `json:"history"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode stats file: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if data.Stats != nil {
		m.stats = data.Stats
	} else {
		m.stats = make(map[string]*GoctionStats)
	}

	if data.History != nil {
		m.history = data.History
	} else {
		m.history = make(map[string][]ExecutionRecord)
	}

	return nil
}

func (m *Manager) save() error {
	file, err := os.Create(m.statsFile)
	if err != nil {
		return fmt.Errorf("failed to create stats file: %w", err)
	}
	defer file.Close()

	data := struct {
		Stats   map[string]*GoctionStats     `json:"stats"`
		History map[string][]ExecutionRecord `json:"history"`
	}{
		Stats:   m.stats,
		History: m.history,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

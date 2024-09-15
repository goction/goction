package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
    dir := filepath.Dir(statsFile)
    if err := os.MkdirAll(dir, 0775); err != nil {
        return nil, fmt.Errorf("failed to create stats directory: %w", err)
    }

    m := &Manager{
        statsFile: statsFile,
        stats:     make(map[string]*GoctionStats),
        history:   make(map[string][]ExecutionRecord),
    }

    // Check if the file exists and is not empty
    if info, err := os.Stat(statsFile); err == nil && info.Size() > 0 {
        if err := m.load(); err != nil {
            return nil, fmt.Errorf("failed to load stats: %w", err)
        }
    } else {
        // If the file doesn't exist or is empty, initialize it
        if err := m.save(); err != nil {
            return nil, fmt.Errorf("failed to initialize stats file: %w", err)
        }
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

	m.stats = data.Stats
	m.history = data.History

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
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode stats: %w", err)
	}

	return nil
}

func (m *Manager) GetAllHistory() map[string][]ExecutionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Créez une copie profonde de l'historique pour éviter les problèmes de concurrence
	allHistory := make(map[string][]ExecutionRecord)
	for name, records := range m.history {
		allHistory[name] = make([]ExecutionRecord, len(records))
		copy(allHistory[name], records)
	}

	return allHistory
}

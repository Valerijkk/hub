package observability

import (
	"sort"
	"sync"
	"time"
)

// OperationMetric — агрегированная статистика по имени операции (OBS-01).
type OperationMetric struct {
	Operation   string  `json:"operation"`
	Count       int64   `json:"count"`
	ErrorCount  int64   `json:"error_count"`
	TotalMs     int64   `json:"total_ms"`
	MaxMs       int64   `json:"max_ms"`
	AvgMs       float64 `json:"avg_ms"`
	LastSeenUTC string  `json:"last_seen_utc,omitempty"`
}

type opAccumulator struct {
	count      int64
	errorCount int64
	totalMs    int64
	maxMs      int64
	lastSeen   time.Time
}

var (
	metricsMu sync.RWMutex
	metrics   = make(map[string]*opAccumulator)
)

func recordOperation(operation string, durationMs int64, failed bool) {
	if operation == "" {
		return
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	acc, ok := metrics[operation]
	if !ok {
		acc = &opAccumulator{}
		metrics[operation] = acc
	}
	acc.count++
	if failed {
		acc.errorCount++
	}
	acc.totalMs += durationMs
	if durationMs > acc.maxMs {
		acc.maxMs = durationMs
	}
	acc.lastSeen = time.Now().UTC()
	recordPrometheusOperation(operation, durationMs, failed)
}

// Snapshot возвращает снимок метрик операций, отсортированный по имени.
func Snapshot() []OperationMetric {
	metricsMu.RLock()
	defer metricsMu.RUnlock()

	out := make([]OperationMetric, 0, len(metrics))
	for name, acc := range metrics {
		m := OperationMetric{
			Operation:  name,
			Count:      acc.count,
			ErrorCount: acc.errorCount,
			TotalMs:    acc.totalMs,
			MaxMs:      acc.maxMs,
		}
		if acc.count > 0 {
			m.AvgMs = float64(acc.totalMs) / float64(acc.count)
		}
		if !acc.lastSeen.IsZero() {
			m.LastSeenUTC = acc.lastSeen.Format(time.RFC3339)
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Operation < out[j].Operation })
	return out
}

// ResetMetrics очищает in-memory метрики (только для тестов).
func ResetMetrics() {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	metrics = make(map[string]*opAccumulator)
	ResetPrometheusMetrics()
}

package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"sync"

	"github.com/leonf08/metrics-yp.git/internal/models"
)

// MemStorage is an in-memory storage for metrics.
type MemStorage struct {
	Storage map[string]models.Metric
	counter int64
	sync.RWMutex
}

// NewStorage creates a new in-memory storage.
func NewStorage() *MemStorage {
	return &MemStorage{
		Storage: make(map[string]models.Metric, 30),
	}
}

// Update updates metrics in the storage.
func (st *MemStorage) Update(_ context.Context, v any) error {
	st.Lock()
	defer st.Unlock()

	m, ok := v.(runtime.MemStats)
	if !ok {
		return errors.New("invalid input data")
	}

	st.Storage = map[string]models.Metric{
		"Alloc":         {Type: "gauge", Val: float64(m.Alloc)},
		"BuckHashSys":   {Type: "gauge", Val: float64(m.BuckHashSys)},
		"Frees":         {Type: "gauge", Val: float64(m.Frees)},
		"GCCPUFraction": {Type: "gauge", Val: m.GCCPUFraction},
		"GCSys":         {Type: "gauge", Val: float64(m.GCSys)},
		"HeapAlloc":     {Type: "gauge", Val: float64(m.HeapAlloc)},
		"HeapIdle":      {Type: "gauge", Val: float64(m.HeapIdle)},
		"HeapInuse":     {Type: "gauge", Val: float64(m.HeapInuse)},
		"HeapObjects":   {Type: "gauge", Val: float64(m.HeapObjects)},
		"HeapReleased":  {Type: "gauge", Val: float64(m.HeapReleased)},
		"HeapSys":       {Type: "gauge", Val: float64(m.HeapSys)},
		"LastGC":        {Type: "gauge", Val: float64(m.LastGC)},
		"Lookups":       {Type: "gauge", Val: float64(m.Lookups)},
		"MCacheInuse":   {Type: "gauge", Val: float64(m.MCacheInuse)},
		"MCacheSys":     {Type: "gauge", Val: float64(m.MCacheSys)},
		"MSpanInuse":    {Type: "gauge", Val: float64(m.MSpanInuse)},
		"MSpanSys":      {Type: "gauge", Val: float64(m.MSpanSys)},
		"Mallocs":       {Type: "gauge", Val: float64(m.Mallocs)},
		"NextGC":        {Type: "gauge", Val: float64(m.NextGC)},
		"NumForcedGC":   {Type: "gauge", Val: float64(m.NumForcedGC)},
		"NumGC":         {Type: "gauge", Val: float64(m.NumGC)},
		"OtherSys":      {Type: "gauge", Val: float64(m.OtherSys)},
		"PauseTotalNs":  {Type: "gauge", Val: float64(m.PauseTotalNs)},
		"StackInuse":    {Type: "gauge", Val: float64(m.StackInuse)},
		"StackSys":      {Type: "gauge", Val: float64(m.StackSys)},
		"Sys":           {Type: "gauge", Val: float64(m.Sys)},
		"TotalAlloc":    {Type: "gauge", Val: float64(m.TotalAlloc)},
	}

	val := rand.Float64()
	st.Storage["RandomValue"] = models.Metric{Type: "gauge", Val: val}

	st.counter++
	st.Storage["PollCount"] = models.Metric{Type: "counter", Val: st.counter}

	return nil
}

// SetVal sets a value for a metric.
func (st *MemStorage) SetVal(_ context.Context, k string, m models.Metric) error {
	st.Lock()
	defer st.Unlock()

	switch m.Type {
	case "gauge":
		st.Storage[k] = m
	case "counter":
		v, ok := st.Storage[k]
		if !ok {
			st.Storage[k] = m
		} else {
			st.Storage[k] = models.Metric{Type: m.Type, Val: v.Val.(int64) + m.Val.(int64)}
		}
	default:
		return errors.New("invalid metric type")
	}

	return nil
}

// GetVal returns a value for a metric.
func (st *MemStorage) GetVal(_ context.Context, k string) (models.Metric, error) {
	st.RLock()
	defer st.RUnlock()

	v, ok := st.Storage[k]
	if !ok {
		return models.Metric{}, fmt.Errorf("metric %s not found", k)
	}

	return v, nil
}

// ReadAll returns all metrics.
func (st *MemStorage) ReadAll(_ context.Context) (map[string]models.Metric, error) {
	st.RLock()
	defer st.RUnlock()
	return st.Storage, nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (st *MemStorage) UnmarshalJSON(data []byte) error {
	s := make(map[string]map[string]models.Metric)

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for k, v := range s["metrics"] {
		if v.Type == "counter" {
			val, ok := v.Val.(float64)
			if !ok {
				return errors.New("failed type assertion")
			}

			st.Storage[k] = models.Metric{Type: v.Type, Val: int64(val)}
		} else {
			st.Storage[k] = v
		}
	}

	return nil
}

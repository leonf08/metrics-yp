package storage

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
)

type Repository interface {
	ReadAll() map[string]interface{}
	Update(interface{})
	SetVal(k string, v interface{}) error
	GetVal(k string) (interface{}, error)
}

type MemStorage struct {
	counter int64
	Storage map[string]interface{} `json:"metrics"`
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Storage: make(map[string]interface{}),
	}
}

func (st *MemStorage) Update(v interface{}) {
	if m, ok := v.(*runtime.MemStats); ok {
		st.Storage = map[string]interface{}{
			"Alloc":         float64(m.Alloc),
			"BuckHashSys":   float64(m.BuckHashSys),
			"Frees":         float64(m.Frees),
			"GCCPUFraction": float64(m.GCCPUFraction),
			"GCSys":         float64(m.GCSys),
			"HeapAlloc":     float64(m.HeapAlloc),
			"HeapIdle":      float64(m.HeapIdle),
			"HeapInuse":     float64(m.HeapInuse),
			"HeapObjects":   float64(m.HeapObjects),
			"HeapReleased":  float64(m.HeapReleased),
			"HeapSys":       float64(m.HeapSys),
			"LastGC":        float64(m.LastGC),
			"Lookups":       float64(m.Lookups),
			"MCacheInuse":   float64(m.MCacheInuse),
			"MCacheSys":     float64(m.MCacheSys),
			"MSpanInuse":    float64(m.MSpanInuse),
			"MSpanSys":      float64(m.MSpanSys),
			"Mallocs":       float64(m.Mallocs),
			"NextGC":        float64(m.NextGC),
			"NumForcedGC":   float64(m.NumForcedGC),
			"NumGC":         float64(m.NumGC),
			"OtherSys":      float64(m.OtherSys),
			"PauseTotalNs":  float64(m.PauseTotalNs),
			"StackInuse":    float64(m.StackInuse),
			"StackSys":      float64(m.StackSys),
			"Sys":           float64(m.Sys),
			"TotalAlloc":    float64(m.TotalAlloc),
		}
	}

	val := rand.Float64()
	st.Storage["RandomValue"] = val

	st.counter++
	st.Storage["PollCount"] = st.counter
}

func (st *MemStorage) SetVal(k string, v interface{}) error {
	switch val := v.(type) {
	case float64:
		st.Storage[k] = val
	case int64:
		_, ok := st.Storage[k]
		if !ok {
			st.Storage[k] = val
			break
		}

		_, ok = st.Storage[k].(int64)
		if !ok {
			cv := st.Storage[k].(float64)
			st.Storage[k] = int64(cv) + val
			break
		}
		
		st.Storage[k] = st.Storage[k].(int64) + val
	default:
		return errors.New("incorrect type of value")
	}

	return nil
}

func (st *MemStorage) GetVal(k string) (interface{}, error) {
	v, ok := st.Storage[k]
	if !ok {
		return nil, fmt.Errorf("metric %s not found", k)
	}

	return v, nil
}

func (st *MemStorage) ReadAll() map[string]interface{} {
	return st.Storage
}

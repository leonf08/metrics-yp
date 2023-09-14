package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
)

type MemStorage struct {
	counter int64                  `json:"-"`
	Storage map[string]interface{} `json:"metrics"`
}

type Metric struct {
	Type string      `json:"type"`
	Val  interface{} `json:"value"`
}

func NewStorage() *MemStorage {
	return &MemStorage{
		Storage: make(map[string]interface{}),
	}
}

func (st *MemStorage) Update(v interface{}) {
	if m, ok := v.(*runtime.MemStats); ok {
		st.Storage = map[string]interface{}{
			"Alloc":         Metric{Type: "gauge", Val: float64(m.Alloc)},
			"BuckHashSys":   Metric{Type: "gauge", Val: float64(m.BuckHashSys)},
			"Frees":         Metric{Type: "gauge", Val: float64(m.Frees)},
			"GCCPUFraction": Metric{Type: "gauge", Val: float64(m.GCCPUFraction)},
			"GCSys":         Metric{Type: "gauge", Val: float64(m.GCSys)},
			"HeapAlloc":     Metric{Type: "gauge", Val: float64(m.HeapAlloc)},
			"HeapIdle":      Metric{Type: "gauge", Val: float64(m.HeapIdle)},
			"HeapInuse":     Metric{Type: "gauge", Val: float64(m.HeapInuse)},
			"HeapObjects":   Metric{Type: "gauge", Val: float64(m.HeapObjects)},
			"HeapReleased":  Metric{Type: "gauge", Val: float64(m.HeapReleased)},
			"HeapSys":       Metric{Type: "gauge", Val: float64(m.HeapSys)},
			"LastGC":        Metric{Type: "gauge", Val: float64(m.LastGC)},
			"Lookups":       Metric{Type: "gauge", Val: float64(m.Lookups)},
			"MCacheInuse":   Metric{Type: "gauge", Val: float64(m.MCacheInuse)},
			"MCacheSys":     Metric{Type: "gauge", Val: float64(m.MCacheSys)},
			"MSpanInuse":    Metric{Type: "gauge", Val: float64(m.MSpanInuse)},
			"MSpanSys":      Metric{Type: "gauge", Val: float64(m.MSpanSys)},
			"Mallocs":       Metric{Type: "gauge", Val: float64(m.Mallocs)},
			"NextGC":        Metric{Type: "gauge", Val: float64(m.NextGC)},
			"NumForcedGC":   Metric{Type: "gauge", Val: float64(m.NumForcedGC)},
			"NumGC":         Metric{Type: "gauge", Val: float64(m.NumGC)},
			"OtherSys":      Metric{Type: "gauge", Val: float64(m.OtherSys)},
			"PauseTotalNs":  Metric{Type: "gauge", Val: float64(m.PauseTotalNs)},
			"StackInuse":    Metric{Type: "gauge", Val: float64(m.StackInuse)},
			"StackSys":      Metric{Type: "gauge", Val: float64(m.StackSys)},
			"Sys":           Metric{Type: "gauge", Val: float64(m.Sys)},
			"TotalAlloc":    Metric{Type: "gauge", Val: float64(m.TotalAlloc)},
		}
	}

	val := rand.Float64()
	st.Storage["RandomValue"] = Metric{Type: "gauge", Val: val}

	st.counter++
	st.Storage["PollCount"] = Metric{Type: "counter", Val: st.counter}
}

func (st *MemStorage) SetVal(k string, v interface{}) error {
	switch val := v.(type) {
	case float64:
		st.Storage[k] = Metric{Type: "gauge", Val: val}
	case int64:
		_, ok := st.Storage[k]
		if !ok {
			st.Storage[k] = Metric{Type: "counter", Val: val}
			break
		}

		m, ok := st.Storage[k].(Metric)
		if !ok {
			return errors.New("failed type assertion")
		}

		c, ok := m.Val.(int64)
		if !ok {
			return errors.New("failed type assertion")
		}

		st.Storage[k] = Metric{Type: "counter", Val: c + val}
	case Metric:
		st.Storage[k] = val
	default:
		return errors.New("incorrect type of value")
	}

	return nil
}

func (st *MemStorage) GetVal(k string) (interface{}, error) {
	v, ok := st.Storage[k]
	if !ok {
		return Metric{}, fmt.Errorf("metric %s not found", k)
	}

	return v, nil
}

func (st *MemStorage) ReadAll() map[string]interface{} {
	return st.Storage
}

func (st *MemStorage) UnmarshalJSON(data []byte) error {
	var s map[string]map[string]Metric

	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for k, v := range s["metrics"] {
		if v.Type == "counter" {
			val, ok := v.Val.(float64)
			if !ok {
				return errors.New("failed type assertion")
			}

			st.Storage[k] = Metric{Type: v.Type, Val: int64(val)}
		} else {
			st.Storage[k] = v
		}
	}

	return nil
}

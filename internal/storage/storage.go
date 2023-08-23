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

type GaugeMetric float64
type CounterMetric int64

type MemStorage struct {
	counter CounterMetric
	storage map[string]interface{}
}

func NewStorage() *MemStorage {
	return &MemStorage{
		storage: make(map[string]interface{}),
	}
}

func (st *MemStorage) Update(v interface{}) {
	if m, ok := v.(*runtime.MemStats); ok {
		st.storage = map[string]interface{} {
			"Alloc": GaugeMetric(m.Alloc),
			"BuckHashSys": GaugeMetric(m.BuckHashSys),
			"Frees": GaugeMetric(m.Frees),
			"GCCPUFraction": GaugeMetric(m.GCCPUFraction),
			"GCSys": GaugeMetric(m.GCSys),
			"HeapAlloc": GaugeMetric(m.HeapAlloc),
			"HeapIdle": GaugeMetric(m.HeapIdle),
			"HeapInuse": GaugeMetric(m.HeapInuse),
			"HeapObjects": GaugeMetric(m.HeapObjects),
			"HeapReleased": GaugeMetric(m.HeapReleased),
			"HeapSys": GaugeMetric(m.HeapSys),
			"LastGC": GaugeMetric(m.LastGC),
			"Lookups": GaugeMetric(m.Lookups),
			"MCacheInuse": GaugeMetric(m.MCacheInuse),
			"MCacheSys": GaugeMetric(m.MCacheSys),
			"MSpanInuse": GaugeMetric(m.MSpanInuse),
			"MSpanSys": GaugeMetric(m.MSpanSys),
			"Mallocs": GaugeMetric(m.Mallocs),
			"NextGC": GaugeMetric(m.NextGC),
			"NumForcedGC": GaugeMetric(m.NumForcedGC),
			"NumGC": GaugeMetric(m.NumGC),
			"OtherSys": GaugeMetric(m.OtherSys),
			"PauseTotalNs": GaugeMetric(m.PauseTotalNs),
			"StackInuse": GaugeMetric(m.StackInuse),
			"StackSys": GaugeMetric(m.StackSys),
			"Sys": GaugeMetric(m.Sys),
			"TotalAlloc": GaugeMetric(m.TotalAlloc),
		}
	}

	val := rand.Float64()
	st.storage["RandomValue"] = GaugeMetric(val)
	
	st.counter++
	st.storage["PollCount"] = st.counter
}

func (st *MemStorage) SetVal(k string, v interface{}) error {
	switch val := v.(type) {
	case float64:
		st.storage[k] = GaugeMetric(val)
	case int64:
		st.storage[k] = CounterMetric(val)
	default:
		return errors.New("Incorrect type of value")
	}

	return nil
}

func (st *MemStorage) GetVal(k string) (interface{}, error) {
	v, ok := st.storage[k]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Metric %s not found", k))
	}

	return v, nil
}

func (st *MemStorage) ReadAll() map[string]interface{} {
	return st.storage
}
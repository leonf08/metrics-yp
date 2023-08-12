package storage

import (
	"math/rand"
	"runtime"
)

type Repository interface {
	GetGaugeMetrics() map[string]GaugeMetric
	GetCounterMetrics() map[string]CounterMetric
	GetGaugeMetricVal(name string) (GaugeMetric, bool)
	GetCounterMetricVal(name string) (CounterMetric, bool)
	WriteGaugeMetric(name string, val float64)
	WriteCounterMetric(name string, val int64)
	UpdateGaugeMetrics()
	UpdateCounterMetrics()
}

type GaugeMetric float64
type CounterMetric int64

type memStorage struct {
	gaugeStorage map[string]GaugeMetric
	counterStorage map[string]CounterMetric
}

func NewStorage() *memStorage {
	return &memStorage{
		gaugeStorage: make(map[string]GaugeMetric),
		counterStorage: make(map[string]CounterMetric),
	}
}

func (mem *memStorage) UpdateGaugeMetrics() {
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	mem.gaugeStorage = map[string]GaugeMetric {
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

	val := rand.Float64()
	mem.gaugeStorage["RandomValue"] = GaugeMetric(val)
}

func (mem *memStorage) UpdateCounterMetrics() {
	mem.counterStorage["PollCount"]++
}

func (mem *memStorage) WriteGaugeMetric(name string, val float64) {
	mem.gaugeStorage[name] = GaugeMetric(val)
}

func (mem *memStorage) WriteCounterMetric(name string, val int64) {
	mem.counterStorage[name] += CounterMetric(val)
}

func (mem memStorage) GetGaugeMetrics() map[string]GaugeMetric {
	return mem.gaugeStorage
}

func (mem memStorage) GetCounterMetrics() map[string]CounterMetric {
	return mem.counterStorage
}

func (mem memStorage) GetGaugeMetricVal(name string) (GaugeMetric, bool) {
	v, ok := mem.gaugeStorage[name]
	return v, ok
}

func (mem memStorage) GetCounterMetricVal(name string) (CounterMetric, bool) {
	v, ok := mem.counterStorage[name]
	return v, ok
}
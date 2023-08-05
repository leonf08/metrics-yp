package storage

import (
	"math/rand"
	"runtime"
)

type gaugeMetric struct {
	metricName string
	value float64
}

type counterMetric struct {
	metricName string
	value int64
}

type MemStorage struct {
	gaugeStorage []gaugeMetric
	counter counterMetric
}

func (mem *MemStorage) UpdateGaugeMetrics(m *runtime.MemStats) {
	mem.gaugeStorage = []gaugeMetric {
		{"Alloc", float64(m.Alloc)},
		{"BuckHashSys", float64(m.BuckHashSys)},
		{"Frees", float64(m.Frees)},
		{"GCCPUFraction", m.GCCPUFraction},
		{"GCSys", float64(m.GCSys)},
		{"HeapAlloc", float64(m.HeapAlloc)},
		{"HeapIdle", float64(m.HeapIdle)},
		{"HeapInuse", float64(m.HeapInuse)},
		{"HeapObjects", float64(m.HeapObjects)},
		{"HeapReleased", float64(m.HeapReleased)},
		{"HeapSys", float64(m.HeapSys)},
		{"LastGC", float64(m.LastGC)},
		{"Lookups", float64(m.Lookups)},
		{"MCacheInuse", float64(m.MCacheInuse)},
		{"MCacheSys", float64(m.MCacheSys)},
		{"MSpanInuse", float64(m.MSpanInuse)},
		{"MSpanSys", float64(m.MSpanSys)},
		{"Mallocs", float64(m.Mallocs)},
		{"NextGC", float64(m.NextGC)},
		{"NumForcedGC", float64(m.NumForcedGC)},
		{"NumGC", float64(m.NumGC)},
		{"OtherSys", float64(m.OtherSys)},
		{"PauseTotalNs", float64(m.PauseTotalNs)},
		{"StackInuse", float64(m.StackInuse)},
		{"StackSys", float64(m.StackSys)},
		{"Sys", float64(m.Sys)},
		{"TotalAlloc", float64(m.TotalAlloc)},
	}

	val := rand.Float64()
	mem.gaugeStorage = append(mem.gaugeStorage, gaugeMetric{"RandomValue", val})
}

func (mem *MemStorage) UpdateCounterMetrics() {
	mem.counter.metricName = "PollCount"
	mem.counter.value++
}

func (mem *MemStorage) WriteGaugeMetric(name string, val float64) {
	mem.gaugeStorage = append(mem.gaugeStorage, gaugeMetric{name, val})
}

func (mem *MemStorage) WriteCounterMetric(val int64) {
	mem.counter.value = val
}

func (mem MemStorage) GetGaugeMetrics() []gaugeMetric {
	return mem.gaugeStorage
}

func (mem MemStorage) GetCounterMetrics() counterMetric {
	return mem.counter
}

func (m gaugeMetric) GetGaugeMetricName() string {
	return m.metricName
}

func (m gaugeMetric) GetGaugeMetricVal() float64 {
	return m.value
}

func (m counterMetric) GetCounterMetricName() string {
	return m.metricName
}

func (m counterMetric) GetCounterMetricVal() int64 {
	return m.value
}
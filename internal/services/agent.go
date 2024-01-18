package services

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
	"strconv"
	"strings"
)

type AgentService struct {
	mode string
	repo Repository
}

func NewAgentService(mode string, repo Repository) *AgentService {
	return &AgentService{
		mode: mode,
		repo: repo,
	}
}

func (a *AgentService) GatherMetrics(ctx context.Context) error {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)
	if err := a.repo.Update(ctx, memStats); err != nil {
		return err
	}

	virtualMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	totalMem := virtualMem.Total
	freeMem := virtualMem.Free

	c, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return err
	}
	cpuUtil := c[0]
	for k, v := range map[string]float64{
		"TotalMemory": float64(totalMem), "FreeMemory": float64(freeMem), "CPUutilization": cpuUtil} {
		if err := a.repo.SetVal(ctx, k, models.Metric{Type: "gauge", Val: v}); err != nil {
			return err
		}
	}

	return nil
}

func (a *AgentService) ReportMetrics(ctx context.Context) ([]string, error) {
	switch a.mode {
	case "json":
		return a.jsonMetrics(ctx)
	case "query":
		return a.queryMetrics(ctx)
	case "batch":
		return a.batchMetrics(ctx)
	default:
		return nil, errors.New("invalid mode")
	}
}

func (a *AgentService) jsonMetrics(ctx context.Context) ([]string, error) {
	str := strings.Builder{}
	gzWriter := gzip.NewWriter(&str)
	defer gzWriter.Close()

	metrics, err := a.repo.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	b := make([]string, 0, len(metrics))
	var m models.MetricJSON
	for k, v := range metrics {
		switch v.Type {
		case "gauge":
			v, ok := v.Val.(float64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			m = models.MetricJSON{
				ID:    k,
				MType: "gauge",
				Value: new(float64),
			}
			*m.Value = v
		case "counter":
			v, ok := v.Val.(int64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			m = models.MetricJSON{
				ID:    k,
				MType: "counter",
				Delta: new(int64),
			}
			*m.Delta = v
		default:
			return nil, errors.New("invalid metric type")
		}

		err = json.NewEncoder(gzWriter).Encode(m)
		if err != nil {
			return nil, err
		}

		err = gzWriter.Flush()
		if err != nil {
			return nil, err
		}
		b = append(b, str.String())

		gzWriter.Reset(&str)
		str.Reset()
	}

	return b, nil
}

func (a *AgentService) queryMetrics(ctx context.Context) ([]string, error) {
	metrics, err := a.repo.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	var valStr string
	b := make([]string, 0, len(metrics))
	for k, v := range metrics {
		switch v.Type {
		case "gauge":
			val, ok := v.Val.(float64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			valStr = strconv.FormatFloat(val, 'f', -1, 64)
		case "counter":
			val, ok := v.Val.(int64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			valStr = strconv.FormatInt(val, 10)
		}

		b = append(b, strings.Join([]string{k, v.Type, valStr}, "/"))
	}

	return b, nil
}

func (a *AgentService) batchMetrics(ctx context.Context) ([]string, error) {
	metrics, err := a.repo.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	b := make([]string, 0, 1)
	m := make([]models.MetricJSON, 0, len(metrics))
	for k, v := range metrics {
		switch v.Type {
		case "gauge":
			val, ok := v.Val.(float64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			m = append(m, models.MetricJSON{
				ID:    k,
				MType: "gauge",
				Value: new(float64),
			})
			*m[len(m)-1].Value = val
		case "counter":
			val, ok := v.Val.(int64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			m = append(m, models.MetricJSON{
				ID:    k,
				MType: "counter",
				Delta: new(int64),
			})
			*m[len(m)-1].Delta = val
		default:
			return nil, errors.New("invalid metric type")
		}
	}

	str := strings.Builder{}
	gzWriter := gzip.NewWriter(&str)
	defer gzWriter.Close()

	err = json.NewEncoder(gzWriter).Encode(m)
	if err != nil {
		return nil, err
	}

	err = gzWriter.Flush()
	if err != nil {
		return nil, err
	}

	b = append(b, str.String())

	return b, nil
}

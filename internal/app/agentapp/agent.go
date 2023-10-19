package agentapp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/auth"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/errorhandling"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Repository interface {
	ReadAll(context.Context) (map[string]any, error)
	Update(context.Context, any) error
	SetVal(context.Context, string, any) error
	GetVal(context.Context, string) (any, error)
}

type Agent struct {
	client  *http.Client
	storage Repository
	logger  logger.Logger
	config  *agentconf.Config
}

func NewAgent(cl *http.Client, st Repository, l logger.Logger, cfg *agentconf.Config) *Agent {
	return &Agent{
		client:  cl,
		storage: st,
		logger:  l,
		config:  cfg,
	}
}

func (a *Agent) Run() error {
	a.logger.Infoln("Running agent")
	m := new(runtime.MemStats)

	pollTime := time.NewTicker(a.config.PollInt)
	reportTime := time.NewTicker(a.config.ReportInt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case <-pollTime.C:
				runtime.ReadMemStats(m)
				if err := a.storage.Update(gCtx, m); err != nil {
					return err
				}
			}
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case <-pollTime.C:
				m, err := mem.VirtualMemoryWithContext(gCtx)
				if err != nil {
					return err
				}
				totalMem := m.Total
				freeMem := m.Free

				c, err := cpu.PercentWithContext(gCtx, 0, false)
				cpuUtil := c[0]
				for n, v := range map[string]float64{
					"TotalMemory": float64(totalMem), "FreeMemory": float64(freeMem), "CPUutilization": cpuUtil} {
					if err := a.storage.SetVal(gCtx, n, v); err != nil {
						return err
					}
				}
			}
		}
	})

	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case <-reportTime.C:
				tasks, err := a.prepareTasks(gCtx)
				if err != nil {
					return err
				}

				pool := newWorkerPool(tasks, runtime.NumCPU())
				pool.run(gCtx)
			}
		}
	})

	return g.Wait()
}

func makeJsonBody(metricName string, value any) (*bytes.Reader, error) {
	var buf bytes.Buffer

	metStruct := new(models.MetricJSON)
	m, ok := value.(storage.Metric)
	if !ok {
		err := errors.New("invalid type assertion")
		return nil, err
	}

	switch v := m.Val.(type) {
	case float64:
		metStruct.ID = metricName
		metStruct.MType = "gauge"
		metStruct.Value = new(float64)
		*metStruct.Value = v
	case int64:
		metStruct.ID = metricName
		metStruct.MType = "counter"
		metStruct.Delta = new(int64)
		*metStruct.Delta = v
	default:
		err := errors.New("invalid metric type")
		return nil, err
	}

	gzWriter := gzip.NewWriter(&buf)
	if err := json.NewEncoder(gzWriter).Encode(&metStruct); err != nil {
		return nil, err
	}

	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func (a *Agent) sendJsonRequest(ctx context.Context, url string, body *bytes.Reader) error {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(body)
	if err != nil {
		return err
	}

	hashSrc := buf.Bytes()

	return errorhandling.Retry(ctx, func() (err error) {
		body.Seek(0, 0)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
		if err != nil {
			a.logger.Errorln(err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		if a.config.IsAuthKeyExists() {
			var hash []byte
			hash, err = auth.CalcHash(hashSrc, []byte(a.config.Key))
			if err != nil {
				a.logger.Errorln(err)
				return
			}

			req.Header.Set("HashSHA256", hex.EncodeToString(hash))
		}

		a.logger.Infoln("Sending request", "address", url)

		resp, err := a.client.Do(req)
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, opErr.Error())
			a.logger.Errorln(err)
			return
		}

		if errors.Is(err, io.EOF) {
			err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, io.EOF.Error())
			a.logger.Errorln(err)
			return
		}

		if err != nil {
			a.logger.Errorln(err)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode > 501 {
			err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, resp.Status)
			a.logger.Errorln(err)
			return
		}

		buf.Reset()
		if _, err = buf.ReadFrom(resp.Body); err != nil {
			a.logger.Errorln(resp.Status, err)
			return
		}

		a.logger.Infoln("Response from the server", "status", resp.Status,
			"body", buf.String())

		buf.Reset()

		return
	})
}

func (a *Agent) prepareTasks(ctx context.Context) ([]*task, error) {
	var tasks []*task
	switch a.config.Mode {
	case "json":
		url := "http://" + a.config.Addr + "/update"

		metrics, err := a.storage.ReadAll(ctx)
		if err != nil {
			a.logger.Errorln(err)
			return nil, err
		}

		bodies := make([]*bytes.Reader, 0)
		for n, v := range metrics {
			b, err := makeJsonBody(n, v)
			if err != nil {
				return nil, err
			}

			bodies = append(bodies, b)
		}

		tasks = make([]*task, len(bodies))
		for i, body := range bodies {
			body := body
			fn := func(ctx context.Context) error {
				return a.sendJsonRequest(ctx, url, body)
			}

			tasks[i] = &task{fn: fn}
		}
	}

	return tasks, nil
}

func (a *Agent) sendMetricBatch(url string) error {
	var buf bytes.Buffer

	metrics, err := a.storage.ReadAll(context.Background())
	if err != nil {
		a.logger.Errorln(err)
		return err
	}

	metSlice := make([]models.MetricJSON, 0)
	for name, value := range metrics {
		metStruct := models.MetricJSON{}
		m, ok := value.(storage.Metric)
		if !ok {
			err := errors.New("invalid type assertion")
			a.logger.Errorln(err)
			return err
		}

		switch v := m.Val.(type) {
		case float64:
			metStruct.ID = name
			metStruct.MType = "gauge"
			metStruct.Value = new(float64)
			*metStruct.Value = v
		case int64:
			metStruct.ID = name
			metStruct.MType = "counter"
			metStruct.Delta = new(int64)
			*metStruct.Delta = v
		default:
			err := errors.New("invalid metric type")
			a.logger.Errorln(err)
			return err
		}

		metSlice = append(metSlice, metStruct)
	}

	gzWriter := gzip.NewWriter(&buf)
	if err := json.NewEncoder(gzWriter).Encode(metSlice); err != nil {
		a.logger.Errorln(err)
		return err
	}

	if err := gzWriter.Close(); err != nil {
		a.logger.Errorln(err)
		return err
	}

	r := bytes.NewReader(buf.Bytes())
	ctx := context.Background()

	err = errorhandling.Retry(ctx, func() (err error) {
		r.Seek(0, 0)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, r)
		if err != nil {
			a.logger.Errorln(err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		a.logger.Infoln("Sending request", "address", url)

		if a.config.IsAuthKeyExists() {
			var hash []byte
			hash, err = auth.CalcHash(buf.Bytes(), []byte(a.config.Key))
			if err != nil {
				a.logger.Errorln(err)
				return
			}

			req.Header.Set("HashSHA256", hex.EncodeToString(hash))
		}

		resp, err := a.client.Do(req)
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, opErr.Error())
			a.logger.Errorln(err)
			return
		}

		if errors.Is(err, io.EOF) {
			err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, io.EOF.Error())
			a.logger.Errorln(err)
			return
		}

		if err != nil {
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode > 501 {
			return errorhandling.ErrRetriable
		}

		buf.Reset()
		if _, err = buf.ReadFrom(resp.Body); err != nil {
			a.logger.Errorln(err)
			return
		}

		a.logger.Infoln("Response from the server", "status", resp.Status,
			"body", buf.String())

		return
	})

	if err != nil {
		a.logger.Errorln(err)
		return err
	}

	buf.Reset()

	return nil
}

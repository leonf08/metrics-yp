package agentapp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

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
	url := "http://" + a.config.Addr + "/update"

	pollTime := time.NewTicker(time.Second * time.Duration(a.config.PollInt))
	reportTime := time.NewTicker(time.Second * time.Duration(a.config.ReportInt))

	for {
		select {
		case <-pollTime.C:
			runtime.ReadMemStats(m)
			if err := a.storage.Update(context.Background(), m); err != nil {
				return err
			}
		case <-reportTime.C:
			if err := a.sendMetricJSON(url); err != nil {
				return err
			}
		}
	}
}

func (a *Agent) sendMetricJSON(url string) error {
	var buf bytes.Buffer

	metrics, err := a.storage.ReadAll(context.Background())
	if err != nil {
		a.logger.Errorln(err)
		return err
	}

	for name, value := range metrics {
		metStruct := new(models.MetricJSON)
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

		gzWriter := gzip.NewWriter(&buf)
		if err := json.NewEncoder(gzWriter).Encode(&metStruct); err != nil {
			a.logger.Errorln(err)
			return err
		}

		if err := gzWriter.Close(); err != nil {
			a.logger.Errorln(err)
			return err
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, &buf)
		if err != nil {
			a.logger.Errorln(err)
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		a.logger.Infoln("Sending request", "address", url)

		var resp *http.Response
		err = errorhandling.Retry(req.Context(), func() error {
			r, err := a.client.Do(req)
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, opErr.Error())
				a.logger.Errorln(err)
				return err
			}

			if err != nil {
				return err
			}

			if r.StatusCode > 501 {
				err = errorhandling.ErrRetriable
			}

			resp = r

			return err
		})

		if err != nil {
			a.logger.Errorln(err)
			return err
		}

		if _, err = buf.ReadFrom(resp.Body); err != nil {
			a.logger.Errorln(err)
			return err
		}

		resp.Body.Close()

		a.logger.Infoln("Response from the server", "status", resp.Status,
			"body", buf.String())

		buf.Reset()
	}

	return nil
}

func (a *Agent) sendMetric(url string) error {
	metrics, err := a.storage.ReadAll(context.Background())
	if err != nil {
		a.logger.Errorln(err)
		return err
	}

	for name, value := range metrics {
		m, ok := value.(storage.Metric)
		if !ok {
			err := errors.New("invalid type assertion")
			a.logger.Errorln(err)
			return err
		}

		switch m.Type {
		case "gauge":
			v, ok := m.Val.(float64)
			if !ok {
				err := errors.New("invalid value type in gauge metric")
				a.logger.Errorln(err)
				return err
			}
			val := strconv.FormatFloat(v, 'f', -1, 64)
			url = strings.Join([]string{url, "gauge", name, val}, "/")
		case "counter":
			v, ok := m.Val.(int64)
			if !ok {
				err := errors.New("invalid value type in counter metric")
				a.logger.Errorln(err)
				return err
			}
			val := strconv.FormatInt(v, 10)
			url = strings.Join([]string{url, "counter", name, val}, "/")
		default:
			err := errors.New("invalid metric type")
			a.logger.Errorln(err)
			return err
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
		if err != nil {
			a.logger.Errorln(err)
			return err
		}

		req.Header.Add("Content-Type", "text/plain")

		a.logger.Infoln("Sending request", req.URL)
		var resp *http.Response
		err = errorhandling.Retry(req.Context(), func() error {
			r, err := a.client.Do(req)
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, opErr.Error())
				a.logger.Errorln(err)
				return err
			}

			if err != nil {
				return err
			}

			defer r.Body.Close()

			if r.StatusCode > 501 {
				err = errorhandling.ErrRetriable
			}

			resp = r

			return err
		})

		if err != nil {
			a.logger.Errorln(err)
			return err
		}

		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if err != nil {
			a.logger.Errorln(err)
			return err
		}
	}

	return nil
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, &buf)
	if err != nil {
		a.logger.Errorln(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	a.logger.Infoln("Sending request", "address", url)
	var resp *http.Response
	err = errorhandling.Retry(req.Context(), func() error {
		r, err := a.client.Do(req)
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			err = fmt.Errorf("%w: %s", errorhandling.ErrRetriable, opErr.Error())
			a.logger.Errorln(err)
			return err
		}

		if err != nil {
			return err
		}

		defer r.Body.Close()

		if r.StatusCode > 501 {
			err = errorhandling.ErrRetriable
		}

		resp = r

		return err
	})
	
	if err != nil {
		a.logger.Errorln(err)
		return err
	}
	defer resp.Body.Close()

	if _, err = buf.ReadFrom(resp.Body); err != nil {
		a.logger.Errorln(err)
		return err
	}

	a.logger.Infoln("Response from the server", "status", resp.Status,
		"body", buf.String())

	buf.Reset()

	return nil
}

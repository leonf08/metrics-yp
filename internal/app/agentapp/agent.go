package agentapp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Agent struct {
	client  *http.Client
	storage storage.Repository
	logger  logger.Logger
	config  *agentconf.Config
}

func NewAgent(cl *http.Client, st storage.Repository, l logger.Logger, cfg *agentconf.Config) *Agent {
	return &Agent{
		client:  cl,
		storage: st,
		logger:  l,
		config:  cfg,
	}
}

func (a *Agent) Run() {
	a.logger.Infoln("Running agent")
	m := new(runtime.MemStats)
	url := "http://" + a.config.Addr + "/update"

	pollTime := time.NewTicker(time.Second * time.Duration(a.config.PollInt))
	reportTime := time.NewTicker(time.Second * time.Duration(a.config.ReportInt))

	for {
		select {
		case <-pollTime.C:
			runtime.ReadMemStats(m)
			a.storage.Update(m)
		case <-reportTime.C:
			a.sendMetricJSON(url)
		}
	}
}

func (a *Agent) sendMetricJSON(url string) {
	var buf bytes.Buffer

	metrics := a.storage.ReadAll()

	for name, value := range metrics {
		metStruct := new(models.Metrics)
		m, ok := value.(storage.Metric)
		if !ok {
			a.logger.Errorln("Invalid type")
			return
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
			a.logger.Errorln("Invalid type of metric, got:", v)
			return
		}

		gzWriter := gzip.NewWriter(&buf)
		if err := json.NewEncoder(gzWriter).Encode(&metStruct); err != nil {
			a.logger.Errorln("Failed to create json", err)
			return
		}

		if err := gzWriter.Close(); err != nil {
			a.logger.Errorln(err)
			return
		}

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, &buf)
		if err != nil {
			a.logger.Errorln("Failed to create http request", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		a.logger.Infoln("Sending request", "address", url)
		resp, err := a.client.Do(req)
		if err != nil {
			a.logger.Errorln("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if _, err = buf.ReadFrom(resp.Body); err != nil {
			a.logger.Errorln("Failed to read response", err)
			return
		}

		a.logger.Infoln("Response from the server", "status", resp.Status,
			"body", buf.String())

		buf.Reset()
	}
}

func (a *Agent) sendMetric(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.config.Timeout))
	defer cancel()

	metrics := a.storage.ReadAll()
	for name, value := range metrics {
		m, ok := value.(storage.Metric)
		if !ok {
			a.logger.Fatalln("invalid element type in storage")
		}

		switch m.Type {
		case "gauge":
			v, ok := m.Val.(float64)
			if !ok {
				a.logger.Fatalln("invalid value type in gauge metric")
			}
			val := strconv.FormatFloat(v, 'f', -1, 64)
			url = strings.Join([]string{url, "gauge", name, val}, "/")
		case "counter":
			v, ok := m.Val.(int64)
			if !ok {
				a.logger.Fatalln("invalid value type in counter metric")
			}
			val := strconv.FormatInt(v, 10)
			url = strings.Join([]string{url, "counter", name, val}, "/")
		default:
			a.logger.Fatalln("invalid metric type")
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			a.logger.Errorln(err)
			break
		}

		req.Header.Add("Content-Type", "text/plain")

		a.logger.Infoln("Sending request", req.URL)
		resp, err := a.client.Do(req)
		if err != nil {
			a.logger.Errorln(err)
			break
		}

		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if err != nil {
			a.logger.Errorln(err)
		}
	}
}

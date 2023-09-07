package agentapp

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Agent struct {
	client  *http.Client
	storage storage.Repository
	config  *agentconf.Config
}

func NewAgent(cl *http.Client, st storage.Repository, cfg *agentconf.Config) *Agent {
	return &Agent{
		client:  cl,
		storage: st,
		config:  cfg,
	}
}

func (a *Agent) Run(log logger.Logger) {
	log.Infoln("Running agent")
	m := new(runtime.MemStats)
	url := "http://" + a.config.Addr + "/update"

	lastPollTime := time.Now()
	lastReportTime := time.Now()

	for {
		currentTime := time.Now()

		if currentTime.Sub(lastPollTime) >= time.Duration(a.config.PollInt*int(time.Second)) {
			lastPollTime = currentTime

			runtime.ReadMemStats(m)
			a.storage.Update(m)
		}

		if currentTime.Sub(lastReportTime) >= time.Duration(a.config.ReportInt*int(time.Second)) {
			lastReportTime = currentTime
			sendMetricJSON(a.client, a.storage, log, url)
		}

		time.Sleep(time.Second)
	}
}

func sendMetricJSON(cl *http.Client, st storage.Repository, log logger.Logger, url string) {
	var buf bytes.Buffer

	metrics := st.ReadAll()

	for name, value := range metrics {
		metStruct := new(models.Metrics)
		m, ok := value.(storage.Metric)
		if !ok {
			log.Errorln("Invalid type")
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
			log.Errorln("Invalid type of metric, got:", v)
			return
		}

		gzWriter := gzip.NewWriter(&buf)
		if err := json.NewEncoder(gzWriter).Encode(&metStruct); err != nil {
			log.Errorln("Failed to create json", err)
			return
		}

		if err := gzWriter.Close(); err != nil {
			log.Errorln(err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, url, &buf)
		if err != nil {
			log.Errorln("Failed to create http request", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		log.Infoln("Sending request", "address", url)
		resp, err := cl.Do(req)
		if err != nil {
			log.Errorln("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if _, err = buf.ReadFrom(resp.Body); err != nil {
			log.Errorln("Failed to read response", err)
			return
		}

		log.Infoln("Response from the server", "status", resp.Status,
			"body", buf.String())

		buf.Reset()
	}
}

// func sendMetric(cl *http.Client, st storage.Repository, log logger.Logger, req string) {
// 	var url string
// 	metrics := st.ReadAll()
// 	for name, value := range metrics {
// 		switch v := value.(type) {
// 		case storage.GaugeMetric:
// 			val := strconv.FormatFloat(float64(v), 'f', -1, 64)
// 			url = strings.Join([]string{req, "gauge", name, val}, "/")
// 		case storage.CounterMetric:
// 			val := strconv.FormatInt(int64(v), 10)
// 			url = strings.Join([]string{req, "counter", name, val}, "/")
// 		default:
// 			log.Errorln("Invalid type of metric, got:", v)
// 			return
// 		}

// 		sendRequest(cl, log, url)
// 	}
// }

// func sendRequest(cl *http.Client, log logger.Logger, url string) {
// 	req, err := http.NewRequest(http.MethodPost, url, nil)
// 	if err != nil {
// 		log.Errorln(err)
// 		return
// 	}

// 	req.Header.Add("Content-Type", "text/plain")

// 	log.Infoln("Sending request", req.URL)
// 	resp, err := cl.Do(req)
// 	if err != nil {
// 		log.Errorln(err)
// 		return
// 	}

// 	_, err = io.Copy(io.Discard, resp.Body)
// 	resp.Body.Close()

// 	if err != nil {
// 		log.Errorln(err)
// 	}
// }

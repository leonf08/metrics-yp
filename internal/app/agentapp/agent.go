package agentapp

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

type Agent struct {
	client *http.Client
	storage storage.Repository
	config *agentconf.Config
}

func NewAgent(cl *http.Client, st storage.Repository, cfg *agentconf.Config) *Agent {
	return &Agent{
		client: cl,
		storage: st,
		config: cfg,
	}
}

func (a *Agent) Run() {
	slog.Info("Running agent")
	m := new(runtime.MemStats)
	request := "http://" + a.config.Addr + "/update"

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
			sendMetricJSON(a.client, a.storage, request)
		}

		time.Sleep(time.Second)
	}
}

func sendMetricJSON(cl *http.Client, st storage.Repository, url string) {
	metrics := st.ReadAll()

	for name, value := range metrics {
		metStruct := new(models.Metrics)
		switch v := value.(type) {
		case storage.GaugeMetric:
			metStruct.ID = name
			metStruct.MType = "gauge"
			metStruct.Value = new(float64)
			*metStruct.Value = float64(v)
		case storage.CounterMetric:
			metStruct.ID = name
			metStruct.MType = "counter"
			metStruct.Delta = new(int64)
			*metStruct.Delta = int64(v)
		default:
			slog.Error("Invalid type of metric, got:", v)
			return
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(&metStruct); err != nil {
			slog.Error("Failed to create json", err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, url, &buf)
		if err != nil {
			slog.Error("Failed to create http request", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		
		slog.Info("Sending request",
				"address", url,
				"body", buf.String())
		resp, err := cl.Do(req)
		if err != nil {
			slog.Error("Failed to send request", err)
			return
		}
		defer resp.Body.Close()

		if _, err = buf.ReadFrom(resp.Body); err != nil {
			slog.Error("Failed to read response", err)
			return
		}

		slog.Info("Response from the server:", 
				"status", resp.Status,
				"body", buf.String())
	}
}

// func sendMetric(cl *http.Client, st storage.Repository, req string) {
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
// 			slog.Error("Invalid type of metric, got:", v)
// 			return
// 		} 
		
// 		sendRequest(cl, url)
// 		time.Sleep(10*time.Millisecond)
// 	}
// }

// func sendRequest(cl *http.Client, url string) {
// 	req, err := http.NewRequest(http.MethodPost, url, nil)
// 	if err != nil {
// 		slog.Error("Failed to create request", err)
// 		return
// 	}

// 	req.Header.Add("Content-Type", "text/plain")

// 	slog.Info("Sending request", req.URL)
// 	resp, err := cl.Do(req)
// 	if err != nil {
// 		slog.Error("Failed to send request", err)
// 		return
// 	}

// 	_, err = io.Copy(io.Discard, resp.Body)
// 	resp.Body.Close()

// 	if err != nil {
// 		slog.Error("", err)
// 	}
// }
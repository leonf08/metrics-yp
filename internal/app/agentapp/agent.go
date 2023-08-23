package agentapp

import (
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/logger"
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

func (a *Agent) Run(log logger.Logger) {
	log.Infoln("Running agent")
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
			sendMetric(a.client, a.storage, log, request)
		}

		time.Sleep(time.Second)
	}
}

func sendMetric(cl *http.Client, st storage.Repository, log logger.Logger, req string) {
	var url string
	metrics := st.ReadAll()
	for name, value := range metrics {
		switch v := value.(type) {
		case storage.GaugeMetric:
			val := strconv.FormatFloat(float64(v), 'f', -1, 64)
			url = strings.Join([]string{req, "gauge", name, val}, "/")
		case storage.CounterMetric:
			val := strconv.FormatInt(int64(v), 10)
			url = strings.Join([]string{req, "counter", name, val}, "/")
		default:
			log.Errorln("Invalid type of metric, got:", v)
			return
		} 
		
		sendRequest(cl, log, url)
		time.Sleep(10*time.Millisecond)
	}
}

func sendRequest(cl *http.Client, log logger.Logger, url string) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Errorln(err)
		return
	}

	req.Header.Add("Content-Type", "text/plain")

	log.Infoln("Sending request", req.URL)
	resp, err := cl.Do(req)
	if err != nil {
		log.Errorln(err)
		return
	}

	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Errorln(err)
	}
}
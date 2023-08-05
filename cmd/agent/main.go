package main

import (
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/storage"
)

const requestForm = "http://localhost:8080/update"

var (
	pollInterval = 2 * time.Second
	reportInterval = 10 * time.Second
	lastPollTime time.Time
	lastReportTime time.Time
)

func main() {
	agentStorage := new(storage.MemStorage)

	for {
		currentTime := time.Now()

		if currentTime.Sub(lastPollTime) >= pollInterval {
			lastPollTime = currentTime
			updateMetrics(agentStorage)	
		}

		if currentTime.Sub(lastReportTime) >= reportInterval {
			lastReportTime = currentTime
			sendGaugeMetric(agentStorage)
			sendCounterMetric(agentStorage)
		}

		time.Sleep(time.Second)
	}
	
}

func updateMetrics(s *storage.MemStorage) {
	metrics := new(runtime.MemStats)
	runtime.ReadMemStats(metrics)

	s.UpdateGaugeMetrics(metrics)
	s.UpdateCounterMetrics()
}

func sendGaugeMetric(s *storage.MemStorage) {
	gaugeMetric := s.GetGaugeMetrics()

	for _, el := range gaugeMetric {
		name := el.GetGaugeMetricName()
		val := strconv.FormatFloat(el.GetGaugeMetricVal(), 'f', -1, 64)
		url := strings.Join([]string{requestForm, "gauge", name, val}, "/")

		resp, err := http.Post(url, "text/plain", nil)
		if err != nil {
			log.Fatalf("Error: %s", err.Error())
		}

		log.Println(resp.Header.Get("Content-Type"))
	}
}

func sendCounterMetric(s *storage.MemStorage) {
	counterMetric := s.GetCounterMetrics()

	name := counterMetric.GetCounterMetricName()
	val := strconv.FormatInt(counterMetric.GetCounterMetricVal(), 64)
	url := strings.Join([]string{requestForm, "counter", name, val}, "/")

	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	log.Println(resp.Header.Get("Content-Type"))
}
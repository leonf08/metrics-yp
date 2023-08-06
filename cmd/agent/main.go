package main

import (
	"fmt"
	"io"
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
	client := &http.Client{}
	
	for {
		currentTime := time.Now()

		if currentTime.Sub(lastPollTime) >= pollInterval {
			lastPollTime = currentTime
			updateMetrics(agentStorage)	
		}

		if currentTime.Sub(lastReportTime) >= reportInterval {
			lastReportTime = currentTime
			sendGaugeMetric(client, agentStorage)
			sendCounterMetric(client, agentStorage)
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

func sendGaugeMetric(cl *http.Client, s *storage.MemStorage) {
	gaugeMetric := s.GetGaugeMetrics()

	for _, el := range gaugeMetric {
		name := el.GetGaugeMetricName()
		val := strconv.FormatFloat(el.GetGaugeMetricVal(), 'f', -1, 64)
		url := strings.Join([]string{requestForm, "gauge", name, val}, "/")

		req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("hello"))
		if err != nil {
			fmt.Println(err)
		}

		req.Header.Add("Content-Type", "text/plain")

		resp, err := cl.Do(req)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(resp.Request.URL)
		fmt.Printf("Status Code: %d\r\n", resp.StatusCode)
		for k, v := range resp.Header {
			fmt.Printf("%s: %v\r\n", k, v)
		}

		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Println(err)
		}
	}
}

func sendCounterMetric(cl *http.Client, s *storage.MemStorage) {
	counterMetric := s.GetCounterMetrics()

	name := counterMetric.GetCounterMetricName()
	val := strconv.FormatInt(counterMetric.GetCounterMetricVal(), 10)
	url := strings.Join([]string{requestForm, "counter", name, val}, "/")

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader("hello"))
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Content-Type", "text/plain")

	resp, err := cl.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Status Code: %d\r\n", resp.StatusCode)
	for k, v := range resp.Header {
		fmt.Printf("%s: %v\r\n", k, v)
	}

	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if err != nil {
		fmt.Println(err)
	}
}
package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/storage"
)

const requestForm = "http://127.0.0.1:8080/update"

var (
	pollInterval = 2 * time.Second
	reportInterval = 10 * time.Second
	lastPollTime time.Time
	lastReportTime time.Time
)

func main() {
	agentStorage := storage.NewStorage()
	client := &http.Client{}
	
	for {
		time.Sleep(time.Second)
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
	}
	
}

func updateMetrics(st storage.Repository) {
	st.UpdateGaugeMetrics()
	st.UpdateCounterMetrics()
}

func sendGaugeMetric(cl *http.Client, st storage.Repository) {
	gaugeMetric := st.GetGaugeMetrics()

	for name, v := range gaugeMetric {
		val := strconv.FormatFloat(float64(v), 'f', -1, 64)
		url := strings.Join([]string{requestForm, "gauge", name, val}, "/")
		sendRequest(cl, url)		
	}
}

func sendCounterMetric(cl *http.Client, st storage.Repository) {
	counterMetric := st.GetCounterMetrics()
	for name, v := range counterMetric {
		val := strconv.FormatInt(int64(v), 10)
		url := strings.Join([]string{requestForm, "counter", name, val}, "/")
		sendRequest(cl, url)
	}
}

func sendRequest(cl *http.Client, url string) {
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/storage"
)

func main() {
	cfg := parseFlags()
	request := "http://" + cfg.Address + "/update"

	agentStorage := storage.NewStorage()
	client := &http.Client{}

	lastPollTime := time.Now()
	lastReportTime := time.Now()
	
	for {
		currentTime := time.Now()

		if currentTime.Sub(lastPollTime) >= time.Duration(cfg.PollInt*int(time.Second)) {
			lastPollTime = currentTime
			updateMetrics(agentStorage)	
		}

		if currentTime.Sub(lastReportTime) >= time.Duration(cfg.ReportInt*int(time.Second)) {
			lastReportTime = currentTime
			sendGaugeMetric(client, agentStorage, request)
			sendCounterMetric(client, agentStorage, request)
		}

		time.Sleep(time.Second)
	}
	
}

func updateMetrics(st storage.Repository) {
	st.UpdateGaugeMetrics()
	st.UpdateCounterMetrics()
}

func sendGaugeMetric(cl *http.Client, st storage.Repository, req string) {
	gaugeMetric := st.GetGaugeMetrics()

	for name, v := range gaugeMetric {
		val := strconv.FormatFloat(float64(v), 'f', -1, 64)
		url := strings.Join([]string{req, "gauge", name, val}, "/")
		sendRequest(cl, url)		
	}
}

func sendCounterMetric(cl *http.Client, st storage.Repository, req string) {
	counterMetric := st.GetCounterMetrics()
	for name, v := range counterMetric {
		val := strconv.FormatInt(int64(v), 10)
		url := strings.Join([]string{req, "counter", name, val}, "/")
		sendRequest(cl, url)
	}
}

func sendRequest(cl *http.Client, url string) {
	defer func() {
		if p := recover(); p != nil {
			log.Fatalf("Panic: %v", p)
		}
	}()

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
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	//"github.com/leonf08/metrics-yp.git/internal/storage"
)

//var serverStorage storage.MemStorage

func GaugeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")[1:]
	if !checkURL(parts) {
		http.Error(w, "Incompleted request", http.StatusNotFound)
		return
	}

	//metricName := parts[2]
	metricVal := parts[3]

	_, err := strconv.ParseFloat(metricVal, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	//serverStorage.WriteGaugeMetric(metricName, val)
}

func CounterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")[1:]
	if !checkURL(parts) {
		http.Error(w, "Incompleted request", http.StatusNotFound)
		return
	}

	metricVal := parts[3]

	_, err := strconv.ParseInt(metricVal, 0, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	//serverStorage.WriteCounterMetric(val)
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func checkURL(p []string) bool {
	if len(p) != 4 {
		return false
	}

	metricName := p[2]

	return metricName != ""
}

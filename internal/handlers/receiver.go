package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

type gauge float64
type counter int64
type MemStorage struct {
	gaugeStorage   map[string]gauge
	counterStorage map[string]counter
}

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

	metricName := parts[2]
	metricVal := parts[3]

	val, err := strconv.ParseFloat(metricVal, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	st := MemStorage{gaugeStorage: make(map[string]gauge)}
	st.gaugeStorage[metricName] = gauge(val)
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

	metricName := parts[2]
	metricVal := parts[3]

	val, err := strconv.ParseInt(metricVal, 0, 64)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	st := MemStorage{counterStorage: make(map[string]counter)}
	st.counterStorage[metricName] = counter(val)
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
	return
}

func checkURL(p []string) bool {
	if len(p) != 4 {
		return false
	}

	metricName := p[2]

	if metricName == "" {
		return false
	}

	return true
}

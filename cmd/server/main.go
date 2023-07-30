package main

import (
	"net/http"

	"github.com/leonf08/metrics-yp.git/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/gauge/", handlers.GaugeHandler)
	mux.HandleFunc("/update/counter/", handlers.CounterHandler)
	mux.HandleFunc("/", handlers.DefaultHandler)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}

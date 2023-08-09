package main

import (
	"flag"
	"time"
)

var (
	addr string
	reportInterval time.Duration
	pollInterval time.Duration
)

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "Host address of the server")
	flag.DurationVar(&reportInterval, "r", 10, "Report interval to server")
	flag.DurationVar(&pollInterval, "p", 2, "Poll interval for metrics")
	flag.Parse()
}

package main

import "flag"

var (
	addr string
	reportInterval int
	pollInterval int
)

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "Host address of the server")
	flag.IntVar(&reportInterval, "r", 10, "Report interval to server")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval for metrics")
	flag.Parse()
}

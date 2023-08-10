package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

var (
	addr string
	reportInterval int
	pollInterval int
)

type Config struct {
	AddressEnv string	`env:"ADDRESS"`
	ReportIntEnv int	`env:"REPORT_INTERVAL"`
	PollIntEnv int		`env:"POLL_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "Host address of the server")
	flag.IntVar(&reportInterval, "r", 10, "Report interval to server")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval for metrics")
	flag.Parse()

	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		log.Fatal(err)
	}

	if conf.AddressEnv != "" {
		addr = conf.AddressEnv
	}

	if conf.ReportIntEnv != 0 {
		reportInterval = conf.ReportIntEnv
	}

	if conf.PollIntEnv != 0 {
		pollInterval = conf.PollIntEnv
	}
}

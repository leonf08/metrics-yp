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

type config struct {
	addressEnv string	`env:"ADDRESS"`
	reportIntEnv int	`env:"REPORT_INTERVAL"`
	pollIntEnv int		`env:"POLL_INTERVAL"`
}

func parseFlags() {
	flag.StringVar(&addr, "a", "localhost:8080", "Host address of the server")
	flag.IntVar(&reportInterval, "r", 10, "Report interval to server")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval for metrics")
	flag.Parse()

	var conf config
	if err := env.Parse(&conf); err != nil {
		log.Fatal(err)
	}


	if conf.addressEnv != "" {
		addr = conf.addressEnv
	}

	if conf.reportIntEnv != 0 {
		reportInterval = conf.reportIntEnv
	}

	if conf.pollIntEnv != 0 {
		pollInterval = conf.pollIntEnv
	}
}

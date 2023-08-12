package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address string	`env:"ADDRESS"`
	ReportInt int	`env:"REPORT_INTERVAL"`
	PollInt int		`env:"POLL_INTERVAL"`
}

func parseFlags() *Config {
	var (
		addr string
		reportInterval int
		pollInterval int
	)

	flag.StringVar(&addr, "a", "localhost:8080", "Host address of the server")
	flag.IntVar(&reportInterval, "r", 10, "Report interval to server")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval for metrics")
	flag.Parse()

	cfg := new(Config)
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Address == "" {
		cfg.Address = addr
	}

	if cfg.ReportInt == 0 {
		cfg.ReportInt = reportInterval
	}

	if cfg.PollInt != 0 {
		cfg.PollInt = pollInterval
	}

	return cfg
}

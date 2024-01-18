package agentconf

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"strconv"
	"strings"
)

const (
	defaultAddress   = "http://127.0.0.1:8080"
	defaultReportInt = 10
	defaultPollInt   = 2
	defaultKey       = ""
	defaultRateLimit = 10
	defaultMode      = "json"
)

type Config struct {
	Addr      string `env:"ADDRESS"`
	ReportInt int    `env:"REPORT_INTERVAL"`
	PollInt   int    `env:"POLL_INTERVAL"`
	Key       string `env:"KEY"`
	RateLim   int    `env:"RATE_LIMIT"`
	Mode      string
}

func MustLoadConfig() Config {
	address := flag.String("a", defaultAddress, "Host address of the server")
	key := flag.String("k", defaultKey, "Authentication key")
	rate := flag.Int("l", defaultRateLimit, "Rate limit for http requests")

	reportInt := defaultReportInt
	flag.Func("r", "Report interval to server", func(s string) (err error) {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}

		reportInt = int(v)

		if reportInt <= 0 {
			return fmt.Errorf("invalid flag value: should be greater than 0, got: %d", reportInt)
		}

		return nil
	})
	pollInt := defaultPollInt
	flag.Func("p", "Poll interval for metrics", func(s string) (err error) {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}

		pollInt = int(v)

		if pollInt <= 0 {
			return fmt.Errorf("invalid flag value: should be greater than 0, got: %d", pollInt)
		}

		return nil
	})
	mode := defaultMode
	flag.Func("m", "Mode of operation, possible values: json (default), batch, query", func(s string) error {
		s = strings.ToLower(s)
		if s != "json" && s != "batch" && s != "query" {
			return fmt.Errorf("invalid flag value, got: %s", s)
		}

		mode = s
		return nil
	})

	flag.Parse()

	cfg := Config{
		Addr:      *address,
		ReportInt: reportInt,
		PollInt:   pollInt,
		Key:       *key,
		RateLim:   *rate,
		Mode:      mode,
	}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return cfg
}

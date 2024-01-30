package agentconf

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress   = "localhost:8080"
	defaultReportInt = 10
	defaultPollInt   = 2
	defaultKey       = ""
	defaultRateLimit = 10
	defaultMode      = "json"
)

// Config is a configuration for the agent
type Config struct {
	// Addr is the address of the server to send metrics to
	Addr string `env:"ADDRESS"`

	// ReportInt is the interval for sending metrics to the server
	ReportInt int `env:"REPORT_INTERVAL"`

	// PollInt is the interval for collecting metrics
	PollInt int `env:"POLL_INTERVAL"`

	// Key used in hash calculation for authentication
	Key string `env:"KEY"`

	// RateLim limits the number of requests per second
	RateLim int `env:"RATE_LIMIT"`

	// Mode of operation
	Mode string
}

// MustLoadConfig loads configuration from environment variables
// and command-line flags. If there is an error, it panics.
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

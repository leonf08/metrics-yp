package agentapp

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/server/logger"
	"github.com/leonf08/metrics-yp.git/internal/storage"
	"go.uber.org/zap"
)

const (
	defaultAddress   = "localhost:8080"
	defaultReportInt = 10
	defaultPollInt   = 2
	defaultKey       = ""
	defaultRateLimit = 10
	defaultMode      = "json"
)

func StartApp() error {
	l, err := initLogger()
	if err != nil {
		return err
	}

	cfg, err := getConfig()
	if err != nil {
		return err
	}

	log := logger.NewLogger(l)

	cl := &http.Client{}
	st := storage.NewStorage()

	agent := NewAgent(cl, st, log, cfg)
	return agent.Run()
}

func initLogger() (logger.Logger, error) {
	lvl, err := zap.ParseAtomicLevel("info")
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zl.Sugar(), nil
}

func getConfig() (*agentconf.Config, error) {
	address := flag.String("a", defaultAddress, "Host address of the server")
	key := flag.String("k", defaultKey, "Authentication key")
	rate := flag.Int("l", defaultRateLimit, "Rate limit for http requests")

	reportInt := defaultReportInt
	flag.Func("r", "Report interval to server", func(s string) (err error) {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return
		}

		reportInt = int(v)

		if reportInt <= 0 {
			return fmt.Errorf("invalid flag value: should be greater than 0, got: %d", reportInt)
		}

		return
	})
	pollInt := defaultPollInt
	flag.Func("p", "Poll interval for metrics", func(s string) (err error) {
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return
		}

		pollInt = int(v)

		if pollInt <= 0 {
			return fmt.Errorf("invalid flag value: should be greater than 0, got: %d", pollInt)
		}

		return
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

	cfg := agentconf.NewConfig(*address, *key, mode, reportInt, pollInt, *rate)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

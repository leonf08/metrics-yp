package agentconf

import "time"

type Config struct {
	Addr      string        `env:"ADDRESS"`
	ReportInt time.Duration `env:"REPORT_INTERVAL"`
	PollInt   time.Duration `env:"POLL_INTERVAL"`
	Key       string        `env:"KEY"`
	RateLim   int           `env:"RATE_LIMIT"`
	Mode      string
}

func NewConfig(addr, key, mode string, reportInt, pollInt, rate int) *Config {
	return &Config{
		Addr:      addr,
		Key:       key,
		ReportInt: time.Duration(reportInt) * time.Second,
		PollInt:   time.Duration(pollInt) * time.Second,
		RateLim:   rate,
		Mode:      mode,
	}
}

func (cfg *Config) IsAuthKeyExists() bool {
	return cfg.Key != ""
}

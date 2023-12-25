package agentconf

type Config struct {
	Addr      string `env:"ADDRESS"`
	ReportInt int    `env:"REPORT_INTERVAL"`
	PollInt   int    `env:"POLL_INTERVAL"`
	Key       string `env:"KEY"`
	RateLim   int    `env:"RATE_LIMIT"`
	Mode      string
}

func NewConfig(addr, key, mode string, reportInt, pollInt, rate int) *Config {
	return &Config{
		Addr:      addr,
		Key:       key,
		ReportInt: reportInt,
		PollInt:   pollInt,
		RateLim:   rate,
		Mode:      mode,
	}
}

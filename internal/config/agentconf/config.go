package agentconf

type Config struct {
	Addr      string `env:"ADDRESS"`
	ReportInt int    `env:"REPORT_INTERVAL"`
	PollInt   int    `env:"POLL_INTERVAL"`
}

func NewConfig(addr string, reportInt, pollInt int) *Config {
	return &Config{
		Addr:      addr,
		ReportInt: reportInt,
		PollInt:   pollInt,
	}
}

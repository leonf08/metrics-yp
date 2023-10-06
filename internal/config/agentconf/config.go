package agentconf

type Config struct {
	Addr      string `env:"ADDRESS"`
	ReportInt int    `env:"REPORT_INTERVAL"`
	PollInt   int    `env:"POLL_INTERVAL"`
	Key       string `env:"KEY"`
}

func NewConfig(addr, key string, reportInt, pollInt int) *Config {
	return &Config{
		Addr:      addr,
		Key: key,
		ReportInt: reportInt,
		PollInt:   pollInt,
	}
}

func (cfg *Config) IsAuthKeyExists() bool {
	return cfg.Key != ""
}
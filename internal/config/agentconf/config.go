package agentconf

type Config struct {
	addr string		`env:"ADDRESS"`
	reportInt int	`env:"REPORT_INTERVAL"`
	pollInt int		`env:"POLL_INTERVAL"`
}

func NewConfig(addr string, reportInt, pollInt int) *Config {
	return &Config{
		addr: addr,
		reportInt: reportInt,
		pollInt: pollInt,
	}
}

func (c Config) Address() string {
	return c.addr
}

func (c Config) ReportInterval() int {
	return c.reportInt
}

func (c Config) Pollinterval() int {
	return c.pollInt
}
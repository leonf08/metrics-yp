package serverconf

type Config struct {
	addr string
}

func NewConfig(addr string) *Config {
	return &Config{
		addr: addr,
	}
}

func (c Config) Address() string {
	return c.addr
}
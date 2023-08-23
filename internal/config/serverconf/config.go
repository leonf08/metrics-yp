package serverconf

type Config struct {
	Addr string `env:"ADDRESS"`
}

func NewConfig(addr string) *Config {
	return &Config{
		Addr: addr,
	}
}

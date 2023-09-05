package serverconf

type Config struct {
	Addr string `env:"ADDRESS"`
	StoreInt int `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore bool `env:"RESTORE"`
}

func NewConfig(storeInt int, addr, filePath string, restore bool) *Config {
	return &Config{
		Addr: addr,
		StoreInt: storeInt,
		FileStoragePath: filePath,
		Restore: restore,
	}
}

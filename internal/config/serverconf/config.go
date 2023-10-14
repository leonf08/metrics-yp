package serverconf

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInt        int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseAddr    string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func NewConfig(storeInt int, addr, filePath, dbAddr, key string, restore bool) *Config {
	return &Config{
		Addr:            addr,
		StoreInt:        storeInt,
		FileStoragePath: filePath,
		DatabaseAddr:    dbAddr,
		Restore:         restore,
		Key:             key,
	}
}

func (cfg *Config) IsInMemStorage() bool {
	return cfg.DatabaseAddr == ""
}

func (cfg *Config) IsFileStorage() bool {
	return cfg.FileStoragePath != "" && cfg.IsInMemStorage()
}

func (cfg *Config) IsAuthKeyExists() bool {
	return cfg.Key != ""
}
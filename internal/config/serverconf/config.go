package serverconf

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInt        int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DataBaseAddr    string `env:"DATABASE_DSN"`
	useDB           bool
}

func NewConfig(storeInt int, addr, filePath, dbAddr string, restore bool) *Config {
	return &Config{
		Addr:            addr,
		StoreInt:        storeInt,
		FileStoragePath: filePath,
		DataBaseAddr:    dbAddr,
		Restore:         restore,
	}
}

func (cfg *Config) UseDB(flag bool) {
	cfg.useDB = flag
}

func (cfg *Config) IsDB() bool {
	return cfg.useDB
}

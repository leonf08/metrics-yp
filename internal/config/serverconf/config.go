package serverconf

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	StoreInt        int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseAddr    string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

func MustLoadConfig() Config {
	address := flag.String("a", ":8080", "Host address of the server")
	storeInt := flag.Int("i", 300, "Store interval for the metrics")
	filePath := flag.String("f", "tmp/metrics-db.json", "Path to file for metrics storage")
	dbAddr := flag.String("d", "", "Database address")
	restore := flag.Bool("r", true, "Load previously saved metrics at the server start")
	key := flag.String("k", "", "Authentication key")
	flag.Parse()

	cfg := Config{
		Addr:            *address,
		StoreInt:        *storeInt,
		FileStoragePath: *filePath,
		Restore:         *restore,
		DatabaseAddr:    *dbAddr,
		Key:             *key,
	}

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return cfg
}

func (cfg Config) IsInMemStorage() bool {
	return cfg.DatabaseAddr == ""
}

func (cfg Config) IsFileStorage() bool {
	return cfg.FileStoragePath != "" && cfg.IsInMemStorage()
}

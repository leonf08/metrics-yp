package serverconf

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

// Config is a struct for server configuration
type Config struct {
	// Addr is the address of the server
	Addr string `env:"ADDRESS"`

	// StoreInt defines the interval for storing metrics if file storage is used
	StoreInt int `env:"STORE_INTERVAL"`

	// FileStoragePath is the path to file where metrics are stored
	FileStoragePath string `env:"FILE_STORAGE_PATH"`

	// Restore defines whether to load previously saved metrics at the server start
	Restore bool `env:"RESTORE"`

	// DatabaseAddr is the address of the database
	DatabaseAddr string `env:"DATABASE_DSN"`

	// SignKey used in hash calculation for authentication
	SignKey string `env:"KEY"`

	// CryptoKey is a path to a file with private key for decryption
	CryptoKey string `env:"CRYPTO_KEY"`
}

// MustLoadConfig loads configuration from environment variables
// and command-line flags. If there is an error, it panics.
func MustLoadConfig() Config {
	address := flag.String("a", ":8080", "Host address of the server")
	storeInt := flag.Int("i", 300, "Store interval for the metrics")
	filePath := flag.String("f", "tmp/metrics-db.json", "Path to file for metrics storage")
	dbAddr := flag.String("d", "", "Database address")
	restore := flag.Bool("r", true, "Load previously saved metrics at the server start")
	key := flag.String("k", "", "Authentication key")
	cryptoKey := flag.String("crypto-key", "", "Path to a file with private key")
	flag.Parse()

	cfg := Config{
		Addr:            *address,
		StoreInt:        *storeInt,
		FileStoragePath: *filePath,
		Restore:         *restore,
		DatabaseAddr:    *dbAddr,
		SignKey:         *key,
		CryptoKey:       *cryptoKey,
	}

	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return cfg
}

// IsInMemStorage returns true if the server is configured to use in-memory storage
func (cfg Config) IsInMemStorage() bool {
	return cfg.DatabaseAddr == ""
}

// IsFileStorage returns true if the server is configured to use additional file storage
func (cfg Config) IsFileStorage() bool {
	return cfg.FileStoragePath != "" && cfg.IsInMemStorage()
}

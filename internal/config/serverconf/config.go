package serverconf

import (
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultAddress       = ":8080"
	defaultStoreInterval = 300
	defaultRestore       = true
)

const (
	flagAddressName       = "address"
	flagStoreIntervalName = "store_interval"
	flagFileStorageName   = "store_file"
	flagDatabaseAddrName  = "database_dsn"
	flagRestoreName       = "restore"
	flagSignKeyName       = "auth_key"
	flagCryptoKeyName     = "crypto_key"
	flagConfigName        = "config"
)

// Config is a struct for server configuration
type Config struct {
	// Addr is the address of the server
	Addr string `env:"ADDRESS"`

	// StoreInt defines the interval for storing metrics if file storage is used
	StoreInt uint `env:"STORE_INTERVAL"`

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
	pflag.StringP(flagAddressName, "a", defaultAddress, "Host address of the server")
	pflag.UintP(flagStoreIntervalName, "i", defaultStoreInterval, "Store interval for the metrics")
	pflag.StringP(flagFileStorageName, "f", "", "Path to file storage")
	pflag.StringP(flagDatabaseAddrName, "d", "", "Database address")
	pflag.BoolP(flagRestoreName, "r", defaultRestore, "Restore metrics from file")
	pflag.StringP(flagSignKeyName, "k", "", "Authentication key")
	pflag.StringP(flagCryptoKeyName, "y", "", "Path to the file with private key")
	pflag.StringP(flagConfigName, "c", "", "Path to the configuration file")

	pflag.Parse()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}

	configFile, ok := os.LookupEnv("CONFIG")
	if !ok {
		configFile = viper.GetString(flagConfigName)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(err)
		}
	}

	address := viper.GetString(flagAddressName)
	fileStoragePath := viper.GetString(flagFileStorageName)
	databaseAddr := viper.GetString(flagDatabaseAddrName)
	restore := viper.GetBool(flagRestoreName)
	storeInt := viper.GetUint(flagStoreIntervalName)
	signKey := viper.GetString(flagSignKeyName)
	cryptoKey := viper.GetString(flagCryptoKeyName)

	cfg := Config{
		Addr:            address,
		StoreInt:        storeInt,
		FileStoragePath: fileStoragePath,
		DatabaseAddr:    databaseAddr,
		Restore:         restore,
		SignKey:         signKey,
		CryptoKey:       cryptoKey,
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

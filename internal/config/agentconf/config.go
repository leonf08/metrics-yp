package agentconf

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	defaultAddress        = "localhost:8080"
	defaultReportInt uint = 10
	defaultPollInt   uint = 2
	defaultRateLimit uint = 10
	defaultMode           = "json"
)

const (
	flagAddressName        = "address"
	flagReportIntervalName = "report_interval"
	flagPollIntervalName   = "poll_interval"
	flagSignKeyName        = "auth_key"
	flagCryptoKeyName      = "crypto_key"
	flagConfigName         = "config"
	flagModeName           = "mode"
	flagRateLimitName      = "rate_limit"
)

type modeEnum string

func (m *modeEnum) Set(s string) error {
	switch s {
	case "json", "batch", "query":
		*m = modeEnum(s)
		return nil
	default:
		return fmt.Errorf("invalid mode value: %s", s)
	}
}

func (m *modeEnum) String() string {
	return string(*m)
}

func (m *modeEnum) Type() string {
	return "modeEnum"
}

// Config is a configuration for the agent
type Config struct {
	// Addr is the address of the server to send metrics to
	Addr string `env:"ADDRESS"`

	// Mode of operation
	Mode string

	// SignKey used in hash calculation for authentication
	SignKey string `env:"KEY"`

	// CryptoKey is a path to a file with public key for encryption
	CryptoKey string `env:"CRYPTO_KEY"`

	// ReportInt is the interval for sending metrics to the server
	ReportInt uint `env:"REPORT_INTERVAL"`

	// PollInt is the interval for collecting metrics
	PollInt uint `env:"POLL_INTERVAL"`

	// RateLim limits the number of requests per second
	RateLim int `env:"RATE_LIMIT"`
}

// MustLoadConfig loads configuration from environment variables
// and command-line flags. If there is an error, it panics.
func MustLoadConfig() Config {
	var mode modeEnum = defaultMode
	pflag.VarP(&mode, flagModeName, "m", "Mode of operation, possible values: json (default), batch, query")

	pflag.StringP(flagAddressName, "a", defaultAddress, "Host address of the server")
	pflag.StringP(flagSignKeyName, "k", "", "Authentication key")
	pflag.UintP(flagRateLimitName, "l", defaultRateLimit, "Rate limit for http requests")
	pflag.StringP(flagCryptoKeyName, "y", "", "Path to a file with public key")
	pflag.UintP(flagReportIntervalName, "r", defaultReportInt, "Report interval to server")
	pflag.UintP(flagPollIntervalName, "p", defaultPollInt, "Poll interval for metrics")
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
	key := viper.GetString(flagSignKeyName)
	cryptoKey := viper.GetString(flagCryptoKeyName)
	reportInt := viper.GetUint(flagReportIntervalName)
	pollInt := viper.GetUint(flagPollIntervalName)
	rate := viper.GetUint(flagRateLimitName)

	cfg := Config{
		Addr:      address,
		Mode:      string(mode),
		SignKey:   key,
		CryptoKey: cryptoKey,
		ReportInt: reportInt,
		PollInt:   pollInt,
		RateLim:   int(rate),
	}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	return cfg
}

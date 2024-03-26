package serverapp

import (
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/httpserver"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
)

// Run starts the application.
// Services, repository, logger, router and server are initialized here.
// Depending on the configuration, file storage is initialized as well.
// The server is started in a separate goroutine.
// The server is stopped by an interrupt signal or an error.
//
// If file storage is enabled, the metrics are restored from the file
// when the server starts. The metrics are saved to the file every period
// of time specified in the configuration.
func Run(cfg serverconf.Config) {
	var (
		r  repo.Repository
		fs services.FileStore
		cr services.Crypto
		ip services.IPChecker
	)

	log := logger.NewLogger()
	s := services.NewHashSigner(cfg.SignKey)

	if cfg.CryptoKey != "" {
		cr = services.NewCryptoService(cfg.CryptoKey)
	}

	if cfg.TrustedSubnet != "" {
		if prefix, err := netip.ParsePrefix(cfg.TrustedSubnet); err == nil {
			ip = services.NewIPChecker(prefix)
		}
	}

	if cfg.IsInMemStorage() {
		r = repo.NewStorage()

		if cfg.IsFileStorage() {
			fileStorage, err := services.NewFileStorage(cfg.FileStoragePath)
			if err != nil {
				log.Error().Err(err).Msg("app - Run - NewFileStorage")
				return
			}
			defer fileStorage.Close()

			if cfg.Restore {
				log.Info().Msg("app - Run - Restore metrics from file")
				err = fileStorage.Load(r)
				if err != nil {
					log.Error().Err(err).Msg("app - Run - fileStorage.Load")
					return
				}
			}

			if cfg.StoreInt > 0 {
				go func() {
					for {
						<-time.After(time.Duration(cfg.StoreInt) * time.Second)
						log.Info().Msg("app - Run - Save metrics to file")
						if err = fileStorage.Save(r); err != nil {
							log.Error().Err(err).Msg("app - Run - fileStorage.Save")
						}
					}
				}()
			} else {
				fs = fileStorage
			}
		}
	} else {
		db, err := repo.NewDB(cfg.DatabaseAddr)
		if err != nil {
			log.Error().Err(err).Msg("app - Run - NewDB")
			return
		}
		defer db.Close()

		r = db
	}

	router := httpserver.NewRouter(s, cr, r, fs, ip, log)
	server := httpserver.NewServer(router, cfg.Addr)
	log.Info().Str("address", cfg.Addr).Msg("app - Run - Starting server")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case err := <-server.Err():
		log.Error().Err(err).Msg("app - Run - server.Err")
	case sig := <-interrupt:
		log.Info().Str("signal", sig.String()).Msg("app - Run - signal")
	}

	log.Info().Msg("app - Run - Shutdown the server")
	err := server.Shutdown()
	if err != nil {
		log.Error().Err(err).Msg("app - Run - server.Shutdown")
	}
}

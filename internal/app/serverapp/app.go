package serverapp

import (
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
	log := logger.NewLogger()

	s := services.NewHashSigner(cfg.SignKey)

	var (
		r  repo.Repository
		fs services.FileStore
		cr services.Crypto
	)
	if cfg.IsInMemStorage() {
		r = repo.NewStorage()

		if cfg.IsFileStorage() {
			fileStorage, err := services.NewFileStorage(cfg.FileStoragePath)
			if err != nil {
				log.Error().Err(err).Msg("app - Run - NewFileStorage")
				return
			}

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

		r = db
	}

	if cfg.CryptoKey != "" {
		cr = services.NewCryptoService(cfg.CryptoKey)
	}

	router := httpserver.NewRouter(s, cr, r, fs, log)
	server := httpserver.NewServer(router, cfg.Addr)
	log.Info().Str("address", cfg.Addr).Msg("app - Run - Starting server")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-server.Err():
		log.Error().Err(err).Msg("app - Run - server.Err")
	case sig := <-interrupt:
		log.Info().Str("signal", sig.String()).Msg("app - Run - signal")
	}

	log.Info().Msg("app - Run - Stopping server")
	if cfg.IsFileStorage() && fs != nil {
		fs.Close()
	}
	err := server.Shutdown()
	if err != nil {
		log.Error().Err(err).Msg("app - Run - server.Shutdown")
	}
}

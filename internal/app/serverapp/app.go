package serverapp

import (
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"github.com/leonf08/metrics-yp.git/internal/httpserver"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(cfg serverconf.Config) {
	log := logger.NewLogger()

	s := services.NewHashSigner(cfg.Key)

	var (
		r  services.Repository
		fs services.FileStore
	)
	if cfg.IsInMemStorage() {
		r = repo.NewStorage()

		if cfg.IsFileStorage() {
			fileStorage, err := services.NewFileStorage(cfg.FileStoragePath)
			if err != nil {
				log.Error("app - Run - NewFileStorage", "error", err)
				return
			}

			if cfg.Restore {
				log.Info("app - Run - Load metrics from file")
				err = fileStorage.Load(r)
				if err != nil {
					log.Error("app - Run - fileStorage.Load", "error", err)
					return
				}
			}

			if cfg.StoreInt > 0 {
				go func() {
					for {
						<-time.After(time.Duration(cfg.StoreInt) * time.Second)
						log.Info("app - Run - Save metrics to file")
						if err = fileStorage.Save(r); err != nil {
							log.Error("app - Run - fileStorage.Save", "error", err)
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
			log.Error("app - Run - NewDB", "error", err)
			return
		}

		r = db
	}

	router := httpserver.NewRouter(s, r, fs, log)
	server := httpserver.NewServer(router, cfg.Addr)
	log.Info("app - Run - server.ListenAndServe", "address", cfg.Addr)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-server.Err():
		log.Error("app - Run - server.Err", "error", err)
	case sig := <-interrupt:
		log.Info("app - Run - interrupt", "signal", sig.String())
	}

	log.Info("app - Run - shutdown")
	err := server.Shutdown()
	if err != nil {
		log.Error("app - Run - server.Shutdown", "error", err)
	}
}

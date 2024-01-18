package agentapp

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/client"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg agentconf.Config) {
	// Init logger, repo, agent, signer
	log := logger.NewLogger()
	r := repo.NewStorage()
	agent := services.NewAgentService(cfg.Mode, r)

	var signer services.Signer
	if cfg.Key != "" {
		signer = services.NewHashSigner(cfg.Key)
	}

	// Create client
	cl := client.NewClient(resty.New(), agent, signer, log, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start client
	log.Info("app - Run - Starting client", "mode", cfg.Mode)
	cl.Start(ctx)
	log.Info("app - Run - Client stopped")
}

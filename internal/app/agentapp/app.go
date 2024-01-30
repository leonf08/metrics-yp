package agentapp

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/client"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
)

// Run runs the agent.
func Run(cfg agentconf.Config) {
	// Init logger, repo, agent, signer
	log := logger.NewLogger()
	r := repo.NewStorage()
	agent := services.NewAgentService(cfg.Mode, r)
	signer := services.NewHashSigner(cfg.Key)

	// Create client
	cl := client.NewClient(resty.New(), agent, signer, log, cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start client
	log.Info().Str("mode", cfg.Mode).Msg("app - Run - Client started")
	cl.Start(ctx)
	log.Info().Msg("app - Run - Client stopped")
}

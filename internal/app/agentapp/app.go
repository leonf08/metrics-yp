package agentapp

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/client/grpc"
	"github.com/leonf08/metrics-yp.git/internal/client/http"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/logger"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"golang.org/x/sync/errgroup"
)

// Run runs the agent.
func Run(cfg agentconf.Config) {
	// Init logger, repo, agent, signer
	log := logger.NewLogger()
	r := repo.NewStorage()
	agent := services.NewAgentService(cfg.Mode, r)
	signer := services.NewHashSigner(cfg.SignKey)

	// Init crypto
	var crypto services.Crypto
	if cfg.CryptoKey != "" {
		crypto = services.NewCryptoService(cfg.CryptoKey)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	g, gtx := errgroup.WithContext(ctx)

	// Create http client
	httpclient := http.NewClient(resty.New(), agent, signer, crypto, log, cfg)

	// Create grpc client
	grpcclient := grpc.NewClient(agent, log, cfg)

	// Start http client
	log.Info().Str("mode", cfg.Mode).Msg("app - Run - HTTP client started")
	g.Go(func() error {
		return httpclient.Start(gtx)
	})

	// Start grpc client
	log.Info().Msg("app - Run - GRPC client started")
	g.Go(func() error {
		return grpcclient.Start(gtx)
	})

	// Wait for the clients to finish
	if err := g.Wait(); err != nil {
		log.Err(err).Msg("app - Run - Error")
	}

	log.Info().Msg("app - Run - Shutdown the client")
}

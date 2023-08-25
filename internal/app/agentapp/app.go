package agentapp

import (
	"flag"
	"net/http"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/storage"
)

func StartApp() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	address := flag.String("a", "localhost:8080", "Host address of the server")
	reportInt := flag.Int("r", 10, "Report interval to server")
	pollInt := flag.Int("p", 2, "Poll interval for metrics")
	flag.Parse()
	
	cfg := agentconf.NewConfig(*address, *reportInt, *pollInt)
	err := env.Parse(cfg)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	storage := storage.NewStorage()

	agent := NewAgent(client, storage, cfg)
	agent.Run()
}
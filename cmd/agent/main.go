package main

import (
	"github.com/leonf08/metrics-yp.git/internal/app/agentapp"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
)

func main() {
	config := agentconf.MustLoadConfig()
	agentapp.Run(config)
}

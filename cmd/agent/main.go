package main

import (
	"fmt"

	"github.com/leonf08/metrics-yp.git/internal/app/agentapp"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
)

var buildVersion, buildDate, buildCommit = "N/A", "N/A", "N/A"

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	config := agentconf.MustLoadConfig()
	agentapp.Run(config)
}

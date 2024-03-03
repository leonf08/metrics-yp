package main

import (
	"fmt"

	"github.com/leonf08/metrics-yp.git/internal/app/serverapp"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
)

var buildVersion, buildDate, buildCommit = "N/A", "N/A", "N/A"

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	config := serverconf.MustLoadConfig()
	serverapp.Run(config)
}

package main

import (
	"github.com/leonf08/metrics-yp.git/internal/app/serverapp"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
)

func main() {
	config := serverconf.MustLoadConfig()
	serverapp.Run(config)
}

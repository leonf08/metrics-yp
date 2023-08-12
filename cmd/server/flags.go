package main

import (
	"flag"
	"os"
)

func parseFlags() string {
	var addr string
	flag.StringVar(&addr, "a", ":8080", "Host address to run server")
	flag.Parse()

	if addrEnv := os.Getenv("ADDRESS"); addrEnv != "" {
		addr = addrEnv
	}

	return addr
}
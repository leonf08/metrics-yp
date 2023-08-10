package main

import (
	"flag"
	"os"
)

var flagAddr string

func parseFlags() {
	flag.StringVar(&flagAddr, "a", ":8080", "Host address to run server")
	flag.Parse()

	if addr := os.Getenv("ADDRESS"); addr != "" {
		flagAddr = addr
	}
}
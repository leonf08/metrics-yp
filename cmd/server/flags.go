package main

import "flag"

var flagAddr string

func parseFlags() {
	flag.StringVar(&flagAddr, "a", ":8080", "Host address to run server")
	flag.Parse()
}
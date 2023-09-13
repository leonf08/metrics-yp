package main

import (
	"github.com/leonf08/metrics-yp.git/internal/app/serverapp"
)

func main() {
	if err := serverapp.StartApp(); err != nil {
		panic(err)
	}
}

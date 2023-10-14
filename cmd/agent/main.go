package main

import (
	"fmt"

	"github.com/leonf08/metrics-yp.git/internal/app/agentapp"
)

func main() {
	if err := agentapp.StartApp(); err != nil {
		fmt.Println(err)
	}
}

package main

import (
	"github.com/leonf08/metrics-yp.git/internal/app/serverapp"
	"github.com/leonf08/metrics-yp.git/internal/config/serverconf"
	"os"
	"runtime"
	"runtime/pprof"
)

func main() {
	fmem, err := os.Create(`profiles/result.pprof`)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()

	runtime.GC() // получаем статистику по использованию памяти

	config := serverconf.MustLoadConfig()
	serverapp.Run(config)

	if err := pprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
}

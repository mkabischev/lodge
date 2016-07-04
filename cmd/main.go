package main

import (
	"flag"
	"log"

	"github.com/mkabischev/logde"
)

func main() {
	bindAddr := flag.String("bind", ":20000", "lodge listen address")
	flag.Parse()

	config := logde.DefaultConfig()
	config.WithAddr(*bindAddr)
	srv, _ := logde.New(config)
	log.Fatal(srv.Run())
}

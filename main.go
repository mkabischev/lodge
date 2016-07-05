package main

import (
	"flag"
	"log"

	"github.com/mkabischev/lodge/server"
)

func main() {
	bindAddr := flag.String("bind", ":20000", "lodge listen address")
	flag.Parse()

	config := server.DefaultConfig()
	config.WithAddr(*bindAddr)
	srv, _ := server.New(config)
	log.Fatal(srv.Run())
}

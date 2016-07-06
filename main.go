package main

import (
	"flag"
	"log"

	"github.com/mkabischev/lodge/server"
	"time"
)

func main() {
	bindAddr := flag.String("bind", ":20000", "lodge listen address")
	flag.Parse()

	srv := server.New(server.NewMemory(10 * time.Second))

	log.Fatal(srv.ListenAndServe(*bindAddr))
}

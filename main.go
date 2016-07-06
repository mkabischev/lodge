package main

import (
	"flag"
	"log"

	"time"

	"github.com/mkabischev/lodge/server"
)

func main() {
	bindAddr := flag.String("bind", ":20000", "lodge listen address")
	usersFile := flag.String("users", "", "Path to users file")
	gcPeriod := flag.Duration("gc_period", 10*time.Second, "Storage cleanup period")
	flag.Parse()

	var users *server.UserList

	if usersFile != nil && *usersFile != "" {
		var err error
		users, err = server.NewUserList(*usersFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	srv := server.New(server.NewMemory(*gcPeriod), users)

	log.Fatal(srv.ListenAndServe(*bindAddr))
}

package main

import (
	"flag"
	"log"

	"github.com/mkabischev/lodge/server"
	"github.com/mkabischev/lodge/server/lru"
)

func main() {
	bindAddr := flag.String("bind", ":20000", "lodge listen address")
	usersFile := flag.String("users", "", "Path to users file")
	buckets := flag.Int("buckets", 100, "Number of buckets")
	bucketSize := flag.Int("bucket_size", 10000, "Number of elements in each bucket")
	flag.Parse()

	var users *server.UserList

	if usersFile != nil && *usersFile != "" {
		var err error
		users, err = server.NewUserList(*usersFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	config := server.DefaultConfig()
	config.Users = users

	storage := server.NewBucketStorage(*buckets, func() server.Storage {
		return server.NewLRUStorage(lru.New(*bucketSize))
	})

	srv := server.New(storage, config)

	log.Fatal(srv.ListenAndServe(*bindAddr))
}

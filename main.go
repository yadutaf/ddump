package main

import (
	"flag"
	"log"

	"github.com/yadutaf/distributed-tcpdump/pkg/server"
)

func main() {
	// Parse arguments
	serverListen := flag.String("listen", ":8475", "Server address")
	flag.Parse()

	// Start the actual application
	log.Fatal(server.Serve(*serverListen))
}

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/yadutaf/distributed-tcpdump/pkg/server"
)

func main() {
	// Parse arguments
	serverListen := flag.String("listen", ":8475", "Server address")
	flag.Parse()

	// Register the worker routes
	http.HandleFunc("/capture", server.PacketCaptureHandler)

	// Start the server
	log.Fatal(server.Serve(*serverListen))
}

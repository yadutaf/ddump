package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/yadutaf/distributed-tcpdump/pkg/server"
)

// openListener Opens a TCP or UNIX listener based one the "listen" string.
// The listen string can be of the forms:
// * For regular TCP sockets: "[HOST:]PORT"
// * For Unix domain sockets: "unix:[SOCKET_PATH]"
func openListener(listen string) (net.Listener, error) {
	// Create listener, with support for Unix domain sockets
	var ln net.Listener
	var err error
	if strings.HasPrefix(listen, "unix:") {
		ln, err = net.Listen("unix", listen[5:])
	} else {
		ln, err = net.Listen("tcp", listen)
	}

	if err != nil {
		return nil, err
	}

	return ln, nil
}

// createSelfFiler creates a pcap filter to exclude ddump-worker itself from the capture
// in case ddump-worker is exposed over TCP. If the TCP socket listens on a specific address,
// only this address is filtered out. For Unix sockets, the function is no-op.
func createSelfFiler(ln net.Listener) (string, error) {
	switch addr := ln.Addr().(type) {
	case *net.TCPAddr:
		if addr.IP.IsUnspecified() {
			return fmt.Sprintf("not (tcp and port %d)", addr.Port), nil
		} else {
			return fmt.Sprintf("not (host %s and tcp and port %d)", addr.IP, addr.Port), nil
		}
	case *net.UnixAddr:
		return "", nil // Nothing to filter out
	default:
		return "", fmt.Errorf("unsupported listener type %T", ln.Addr())
	}
}

func main() {
	// Parse arguments
	serverListen := flag.String("listen", ":8475", "Server address")
	flag.Parse()

	// Create listener
	ln, err := openListener(*serverListen)
	if err != nil {
		log.Fatalf("Failed to open listener '%s': %v", *serverListen, err)
	}
	log.Printf("Listening on %s", *serverListen)

	// Create server, exclude listener from the capture
	filterOutSelf, err := createSelfFiler(ln)
	if err != nil {
		log.Fatalf("Failed to create self-filter %v", err)
	}
	pcapServer := server.NewPcapServer(filterOutSelf)

	// Register the worker routes
	http.HandleFunc("/capture", pcapServer.PacketCaptureHandler)

	// Start the server
	log.Fatal(http.Serve(ln, nil))
}

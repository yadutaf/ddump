package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yadutaf/distributed-tcpdump/pkg/merger"
	"github.com/yadutaf/distributed-tcpdump/pkg/streamer"
)

func main() {
	// Parse command line arguments
	caCertPath := flag.String("ca", "", "Path to trusted CA store. Defaults to system CA")
	clientCertPath := flag.String("cert", "", "Path to PEM encoded client certificate. Defaults to none")
	clientKeyPath := flag.String("key", "", "Path to PEM encoded client key. Defaults to none")
	flag.Parse()

	targetUrls := flag.Args()

	if len(targetUrls) < 1 {
		log.Printf("Missing arguments")
		flag.Usage()
		os.Exit(1)
	}

	if len(*clientCertPath) > 0 && len(*clientKeyPath) == 0 {
		log.Fatalf("--key is mandatory when configuring --cert")
	}

	if len(*clientCertPath) == 0 && len(*clientKeyPath) > 0 {
		log.Fatalf("--cert is mandatory when configuring --key")
	}

	// Create TLS client configuration
	tlsClientConfig, err := streamer.InitTlSConfig(*caCertPath, *clientCertPath, *clientKeyPath)
	if err != nil {
		log.Fatalf("Failed to configure TLS client: %v", err)
	}

	// Create the stream merger
	pcapStreamMerger := merger.NewPcapStreamMerger(os.Stdout)

	// Open and register streams
	childStreamer := streamer.NewStreamer(tlsClientConfig)
	for _, targetUrl := range targetUrls {
		resp, err := childStreamer.OpenStream(targetUrl)
		if err != nil {
			log.Fatalf("streamer.OpenStream(): %v", err)
		}
		pcapStreamMerger.Add(resp)
	}

	// Wait for signal in the background
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Printf("Exiting...")
		pcapStreamMerger.Close()
	}()

	// Start the stream merge
	if err := pcapStreamMerger.Start(); err != nil {
		log.Fatalf("Failed to merge the pcap streams: %v", err)
	}

	// All done !
	log.Printf("All done!")
}

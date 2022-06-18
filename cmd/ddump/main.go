package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yadutaf/distributed-tcpdump/pkg/merger"
)

func main() {
	// Parse command line arguments
	flag.Parse()
	targetUrls := flag.Args()

	if len(targetUrls) < 1 {
		log.Printf("Missing arguments")
		flag.Usage()
		os.Exit(1)
	}

	// Create the stream merger
	pcapStreamMerger := merger.NewPcapStreamMerger(os.Stdout)

	// Open and register streams
	for _, targetUrl := range targetUrls {
		resp, err := merger.OpenChildStream(targetUrl)
		if err != nil {
			log.Fatalf("streamer.OpenChildStream(): %v", err)
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

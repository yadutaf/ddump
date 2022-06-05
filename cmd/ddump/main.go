package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"

	"github.com/yadutaf/distributed-tcpdump/pkg/capture"
	"github.com/yadutaf/distributed-tcpdump/pkg/streamer"
)

// fanIn merges all pkt from all streams. Cancellation is supposed to e supported by producers
func fanIn(capturedPacketsStreams ...<-chan CapturedPacket) <-chan CapturedPacket {
	var wg sync.WaitGroup
	out := make(chan CapturedPacket)

	// Packet forwarder
	output := func(c <-chan CapturedPacket) {
		defer wg.Done()

		for pkt := range c {
			out <- pkt
		}
	}

	// Start a packet forwarder for each input stream
	for _, c := range capturedPacketsStreams {
		go output(c)
		wg.Add(1)
	}

	// When all forwarders are done, close donwstream too
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	// Parse comand line arguments
	flag.Parse()
	targetUrls := flag.Args()

	if len(targetUrls) < 1 {
		log.Printf("Missing arguments")
		flag.Usage()
		os.Exit(1)
	}

	// Create the cancel signal
	done := make(chan struct{})

	// Open streams
	respStreams := []<-chan CapturedPacket{}

	for _, targetUrl := range targetUrls {
		resp, err := streamer.OpenChildStream(targetUrl)
		if err != nil {
			log.Fatalf("http.Get(): %v", err)
		}
		respStreams = append(respStreams, StartPcapStreamer(done, resp))
	}

	// Start the FanIn
	packetFanIn := fanIn(respStreams...)

	// Prepare the exit path
	defer func() {
		// Signal we are exiting
		log.Printf("Exiting...")
		close(done)

		// Drain the queue, discard any pending packet (i.e. without writing to closed output)
		log.Printf("Draining...")
		for range packetFanIn {
		}
		log.Printf("All done...")
	}()

	// Initialize packet writer
	pcapw := pcapgo.NewWriter(os.Stdout)
	if err := pcapw.WriteFileHeader(capture.MAX_PACKET_LENGTH, layers.LinkTypeLinuxSLL); err != nil {
		log.Fatalf("pcap.WriteFileHeader(): %v", err)
	}

	// Cleanly exit on SIGINT
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Forward packets
	for {
		select {
		case capturedPacket, ok := <-packetFanIn:
			if !ok {
				// Downstream channel is closed
				log.Printf("Upstream channel is closed")
				return
			}
			// Forward pending packets
			log.Printf("Fowarding packets")
			if err := pcapw.WritePacket(capturedPacket.ci, capturedPacket.data); err != nil {
				log.Printf("Downstream seems to be closed, exiting: %v", err)
				return
			}
		case <-sigs:
			return
		}
	}
}

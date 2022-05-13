package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

const (
	MAX_PACKET_LENGTH         = 9000
	DEFAULT_CAPTURE_INTERFACE = "any"
	DEFAULT_CAPTURE_FILTER    = ""
)

func capture(ifName string, filter string, w io.Writer) error {
	handle, err := pcap.OpenLive(ifName, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("pcap.OpenLive(): %v", err)
	}

	if err := handle.SetBPFFilter(filter); err != nil {
		return fmt.Errorf("handle.SetBPFFilter(): %v", err)
	}

	pcapw := pcapgo.NewWriter(w)
	if err := pcapw.WriteFileHeader(MAX_PACKET_LENGTH, handle.LinkType()); err != nil {
		return fmt.Errorf("WriteFileHeader: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := pcapw.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
			return fmt.Errorf("pcap.WritePacket(): %v", err)
		}
	}

	return nil
}

func httpCaptureHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request
	captureIfName := r.URL.Query().Get("interface")
	if captureIfName == "" {
		captureIfName = DEFAULT_CAPTURE_INTERFACE
	}
	captureFilter := r.URL.Query().Get("filter")
	if captureFilter == "" {
		captureFilter = DEFAULT_CAPTURE_FILTER
	}

	// Set content type
	w.Header().Set("Content-Type", "application/vnd.tcpdump.pcap")

	// Capture
	if err := capture(captureIfName, captureFilter, w); err != nil {
		log.Printf("capture: %v", err)
		return
	}
}

func main() {
	// Parse arguments
	serverListen := flag.String("listen", ":8475", "Server address")
	flag.Parse()

	http.HandleFunc("/capture", httpCaptureHandler)

	log.Printf("Listening on %s", *serverListen)
	log.Fatal(http.ListenAndServe(*serverListen, nil))
}

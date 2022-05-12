package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

const (
	MAX_PACKET_LENGTH = 9000
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

func openDestination(destination string) (*os.File, error) {
	if destination == "-" {
		if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
			return nil, fmt.Errorf("Destination is a terminal. Please use -destination or redirect the output")
		}
		return os.Stdout, nil
	}

	return os.Create(destination)
}

func main() {
	// Parse arguments
	captureIfName := flag.String("interface", "any", "Name of the interface to capture")
	captureFilter := flag.String("filter", "", "capture filter using tcpdump's DSL")
	captureDestination := flag.String("destination", "-", "destination file for captured packets")
	flag.Parse()

	// Configure destination
	f, err := openDestination(*captureDestination)
	if err != nil {
		log.Fatalf("Failed to set destination to %s: %v", *captureDestination, err)
	}
	defer f.Close()

	// Capture
	if err := capture(*captureIfName, *captureFilter, f); err != nil {
		log.Fatalf("capture: %v", err)
	}
}

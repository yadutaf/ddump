// Package capture implements the main capture logic. It is meant to be
// re-usable and use only standard interfaces.
//
// Example usage:
//  package main
//
//  // Importing fmt and time
//  import (
//      "time"
//
//      "github.com/yadutaf/ddump/pkg/capture"
//  )
//
//  func main() {
//      // Prepare the capture file
//      f, err := os.Create("/tmp/capture.pcap")
//      if err != nil {
//          panic(e)
//      }
//
//      // Start the capture in the background
//      done := make(chan struct{})
//      go func() {
//          err := capture.Capture("eth0", "tcp port 80")
//          if err != nil {
//                panic(e)
//          }
//          close(done)
//      }
//
//      // Stop the capture after 3s, or early exit
//      select {
//      case <-time.After(3 * time.Second):
//      case <-done:
//      }
//
//      // Close the writer
//      f.Close()
//
//      // Wait until done
//      <-done
//  }
package capture

import (
	"fmt"
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

const (
	MAX_PACKET_LENGTH = 9000 // Maximum packet size supported by ddump, in bytes. 9000 is a "Jumbo frame"
)

// Capture is a long running function. It starts a live capture session on
// interface ifName for all packets matching the provided filter. Matching
// packets are copied to the w writer. Note: ifName can be the name of an
// actual interface or the special "any" to capture from all interfaces.
//
// This function exits when a write error occurs. For example, if the writer
// is a TCP connection, this function will exit when the downstream connection
// is closed. Conversly, a caller may close the writer to request the capture
// end.
func Capture(ifName string, filter string, w io.Writer) error {
	handle, err := pcap.OpenLive(ifName, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("pcap.OpenLive(): %v", err)
	}

	if err := handle.SetBPFFilter(filter); err != nil {
		return fmt.Errorf("handle.SetBPFFilter(): %v", err)
	}

	pcapw := pcapgo.NewWriter(w)
	if err := pcapw.WriteFileHeader(MAX_PACKET_LENGTH, handle.LinkType()); err != nil {
		return fmt.Errorf("pcap.WriteFileHeader(): %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := pcapw.WritePacket(packet.Metadata().CaptureInfo, packet.Data()); err != nil {
			break // Downstream closed, exit
		}
	}

	return nil
}

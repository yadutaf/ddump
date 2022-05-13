package capture

import (
	"fmt"
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

const (
	MAX_PACKET_LENGTH = 9000
)

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
			return fmt.Errorf("pcap.WritePacket(): %v", err)
		}
	}

	return nil
}

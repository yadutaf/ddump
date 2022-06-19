package merger

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"

	"github.com/yadutaf/ddump/pkg/capture"
)

// capturedPacket is a thin wrapper to keep the packet header and payload together
type capturedPacket struct {
	data []byte
	ci   gopacket.CaptureInfo
}

// PcapStreamMerger encapsultes the logic behind merging input streams
type PcapStreamMerger struct {
	inStreams []<-chan capturedPacket
	outStream io.Writer
	done      chan struct{}
}

func NewPcapStreamMerger(outStream io.Writer) *PcapStreamMerger {
	return &PcapStreamMerger{
		inStreams: []<-chan capturedPacket{},
		outStream: outStream,
		done:      make(chan struct{}),
	}
}

// Add registers a new stream in the merger.
func (m *PcapStreamMerger) Add(inStream io.ReadCloser) {
	// Initialize the channels
	packetStream := make(chan capturedPacket)
	streamerDone := make(chan struct{})

	// Start the streamer in the background
	go func() {
		defer close(packetStream)
		defer close(streamerDone)

		// When done, close the input stream to signal we are done
		// When exiting naturally, close the input stream to cleanup
		go func() {
			select {
			case <-m.done:
			case <-streamerDone:
			}
			log.Printf("Closing reader to stop streamer")
			inStream.Close()
		}()

		// Initialize packet reader
		pcapReader, err := pcapgo.NewReader(inStream)
		if err != nil {
			log.Fatalf("pcapgo.NewReader(): %v", err)
		}

		// Stream received packets
		for {
			data, ci, err := pcapReader.ZeroCopyReadPacketData()
			if err != nil {
				log.Printf("Upstream error, exiting")
				break
			}

			packetStream <- capturedPacket{data: data, ci: ci}
		}

		log.Printf("Streamer is exiting")

	}()

	// Register the streamer
	m.inStreams = append(m.inStreams, packetStream)
}

// Start the pcap stream merge process
func (m *PcapStreamMerger) Start() error {
	// Initialize packet writer
	pcapw := pcapgo.NewWriter(m.outStream)
	if err := pcapw.WriteFileHeader(capture.MAX_PACKET_LENGTH, layers.LinkTypeLinuxSLL); err != nil {
		return fmt.Errorf("pcap.WriteFileHeader(): %v", err)
	}

	// Start the FanIn
	packetFanIn := fanIn(m.inStreams...)

	// Make sure we drain the merged stream on exit
	defer func() {
		log.Printf("Draining merged stream")
		for {
			if _, ok := <-packetFanIn; !ok {
				break
			}
		}
	}()

	// Forward packets
	for {
		select {
		case capturedPacket, ok := <-packetFanIn:
			// Is downstream channel closed ?
			if !ok {
				log.Printf("Upstream channel is closed")
				return nil
			}

			// Forward pending packets
			log.Printf("Fowarding packets")
			if err := pcapw.WritePacket(capturedPacket.ci, capturedPacket.data); err != nil {
				log.Printf("Downstream seems to be closed, exiting: %v", err)
				return nil
			}
		case <-m.done:
			// Drain any buffered packets
			for range packetFanIn {
			}
			return nil
		}
	}
}

func (m *PcapStreamMerger) Close() {
	close(m.done)
}

// fanIn merges all pkt from all streams. Cancellation is supposed to e supported by producers
func fanIn(capturedPacketsStreams ...<-chan capturedPacket) <-chan capturedPacket {
	var wg sync.WaitGroup
	out := make(chan capturedPacket)

	// Packet forwarder
	output := func(c <-chan capturedPacket) {
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

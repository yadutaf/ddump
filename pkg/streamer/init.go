package streamer

import (
	"fmt"
	"io"
	"log"
	"net/url"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcapgo"
)

// CapturedPacket is a thin wrapper to keep the packet header and payload together
type CapturedPacket struct {
	data []byte
	ci   gopacket.CaptureInfo
}

// OpenChildStream is a factory for streamer. Based on the protocol, it instanciate the appropriate streamer
func OpenChildStream(targetUrl string) (io.ReadCloser, error) {
	// Parse targetUrl
	u, err := url.Parse(targetUrl)
	if err != nil {
		return nil, err
	}

	// Select child streamer based on te scheme
	switch u.Scheme {
	case "http", "https":
		return openChildHTTPStream(u)
	default:
		return nil, fmt.Errorf("unsupported scheme '%s' in target url '%s'", u.Scheme, targetUrl)
	}
}

// StartPcapStreamer starts the streamer in the background. It expects a raw
// pcap stream as input and consumes it until EOF, error or done. On stream
// end, the input stream is automatically closed.
// Captured packets are streamed over the returned channel
func StartPcapStreamer(done <-chan struct{}, r io.ReadCloser) <-chan CapturedPacket {
	out := make(chan CapturedPacket)
	streamerDone := make(chan struct{})

	go func() {
		defer close(out)
		defer close(streamerDone)

		// When done, close the input stream to signal we are done
		// When exiting naturally, close the input stream to cleanup
		go func() {
			select {
			case <-done:
			case <-streamerDone:
			}
			log.Printf("Closing reader to stop streamer")
			r.Close()
		}()

		// Initialize packet reader
		pcapReader, err := pcapgo.NewReader(r)
		if err != nil {
			log.Fatalf("pcapgo.NewReader(): %v", err)
		}

		// Stream received packets
		for {
			data, ci, err := pcapReader.ZeroCopyReadPacketData()
			if err != nil {
				log.Printf("Upstream error, exiing")
				break
			}

			out <- CapturedPacket{data: data, ci: ci}
		}

		log.Printf("Streamer is exiting")

	}()

	return out
}

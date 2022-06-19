package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yadutaf/distributed-tcpdump/pkg/capture"
)

const (
	DEFAULT_CAPTURE_FILTER = ""
)

type FlushedResponseWriter struct {
	w    *http.ResponseWriter
	done chan struct{}
}

type PcapServer struct {
	additionalFiler string
}

func NewFlushedResponseWriter(w *http.ResponseWriter) *FlushedResponseWriter {
	frw := FlushedResponseWriter{
		w:    w,
		done: make(chan struct{}),
	}

	// Flush every second if bytes were sent and until done
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if f, ok := (*frw.w).(http.Flusher); ok {
					f.Flush()
				}
			case <-frw.done:
				return
			}
		}
	}()

	return &frw
}

func (frw *FlushedResponseWriter) Close() {
	close(frw.done)
}

func (frw *FlushedResponseWriter) Write(p []byte) (n int, err error) {
	return (*frw.w).Write(p)
}

func NewPcapServer(additionalFiler string) *PcapServer {
	return &PcapServer{
		additionalFiler: additionalFiler,
	}
}

func (s *PcapServer) buildFilter(captureFilter string) string {
	// Apply additional filters (typically to filter ourselves out, to avoid amplification)
	if captureFilter == "" {
		return s.additionalFiler
	} else if s.additionalFiler != "" {
		return fmt.Sprintf("(%s) and (%s)", captureFilter, s.additionalFiler)
	} else {
		return captureFilter
	}
}

func (s *PcapServer) PacketCaptureHandler(w http.ResponseWriter, r *http.Request) {
	// Decode request
	captureFilter := r.URL.Query().Get("filter")
	if captureFilter == "" {
		captureFilter = DEFAULT_CAPTURE_FILTER
	}

	// Apply any additional capture filters
	captureFilter = s.buildFilter(captureFilter)

	// Set content type
	w.Header().Set("Content-Type", "application/vnd.tcpdump.pcap")

	// Capture
	pcapWriter := NewFlushedResponseWriter(&w)
	defer pcapWriter.Close()
	if err := capture.Capture("any", captureFilter, pcapWriter); err != nil {
		log.Printf("capture: %v", err)
		return
	}
}

package server

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/yadutaf/distributed-tcpdump/pkg/capture"
)

const (
	DEFAULT_CAPTURE_INTERFACE = "any"
	DEFAULT_CAPTURE_FILTER    = ""
)

type FlushedResponseWriter struct {
	w    *http.ResponseWriter
	done chan struct{}
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

func PacketCaptureHandler(w http.ResponseWriter, r *http.Request) {
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
	pcapWriter := NewFlushedResponseWriter(&w)
	defer pcapWriter.Close()
	if err := capture.Capture(captureIfName, captureFilter, pcapWriter); err != nil {
		log.Printf("capture: %v", err)
		return
	}
}

func Serve(listen string) error {
	// Create listener, with support for Unix domain sockets
	var ln net.Listener
	var err error
	if strings.HasPrefix(listen, "unix:") {
		ln, err = net.Listen("unix", listen[5:])
	} else {
		ln, err = net.Listen("tcp", listen)
	}

	if err != nil {
		return err
	}
	log.Printf("Listening on %s", listen)

	// Start server
	return http.Serve(ln, nil)
}

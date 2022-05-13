package server

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/yadutaf/distributed-tcpdump/pkg/capture"
)

const (
	DEFAULT_CAPTURE_INTERFACE = "any"
	DEFAULT_CAPTURE_FILTER    = ""
)

type flushedResponseWriter struct {
	w *http.ResponseWriter
}

func (frw *flushedResponseWriter) Write(p []byte) (n int, err error) {
	n, err = (*frw.w).Write(p)
	if err != nil {
		return
	}

	if f, ok := (*frw.w).(http.Flusher); ok {
		f.Flush()
	}

	return
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
	pcapWriter := flushedResponseWriter{w: &w}
	if err := capture.Capture(captureIfName, captureFilter, &pcapWriter); err != nil {
		log.Printf("capture: %v", err)
		return
	}
}

func Serve(listen string) error {
	http.HandleFunc("/capture", httpCaptureHandler)

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

package server

import (
	"log"
	"net/http"

	"github.com/yadutaf/distributed-tcpdump/pkg/capture"
)

const (
	DEFAULT_CAPTURE_INTERFACE = "any"
	DEFAULT_CAPTURE_FILTER    = ""
)

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
	if err := capture.Capture(captureIfName, captureFilter, w); err != nil {
		log.Printf("capture: %v", err)
		return
	}
}

func Serve(listen string) error {
	http.HandleFunc("/capture", httpCaptureHandler)

	log.Printf("Listening on %s", listen)
	return http.ListenAndServe(listen, nil)
}

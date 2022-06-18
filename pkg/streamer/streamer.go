package streamer

import (
	"crypto/tls"
	"io"
	"net/http"
)

type Streamer struct {
	client *http.Client
}

func NewStreamer(tlsClientConfig *tls.Config) *Streamer {
	// Create the client
	return &Streamer{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsClientConfig,
			},
		},
	}
}

// OpenStream opens a connection with the target ddump-worker
func (s *Streamer) OpenStream(targetUrl string) (io.ReadCloser, error) {
	resp, err := s.client.Get(targetUrl)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

package merger

import (
	"io"
	"net/http"
)

// OpenChildStream opens a connection with the target ddump-worker
func OpenChildStream(targetUrl string) (io.ReadCloser, error) {
	resp, err := http.Get(targetUrl)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

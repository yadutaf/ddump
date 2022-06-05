package streamer

import (
	"io"
	"net/http"
	"net/url"
)

func openChildHTTPStream(u *url.URL) (io.ReadCloser, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

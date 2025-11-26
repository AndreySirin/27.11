package client

import (
	"fmt"
	"net/http"
	"time"
)

const (
	statusAvailable    = "available"
	statusNotAvailable = "not available"
)

type Client struct {
	*http.Client
}

func New() *Client {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	return &Client{
		Client: client,
	}
}
func (c *Client) CheckLink(url string) (string, error) {

	headResponse, err := c.Head(url)
	if err != nil {
		return statusNotAvailable, fmt.Errorf("unable to check link: %w", err)
	}
	defer headResponse.Body.Close()

	if headResponse.StatusCode == http.StatusNotImplemented ||
		headResponse.StatusCode == http.StatusMethodNotAllowed {
		r, errCreateReq := http.NewRequest(http.MethodGet, url, nil)
		if errCreateReq != nil {
			return statusNotAvailable, fmt.Errorf("could not create request: %w", errCreateReq)
		}
		r.Header.Set("Range", "bytes=0-0")
		getResponse, errDoRequest := c.Do(r)
		if errDoRequest != nil {
			return statusNotAvailable, fmt.Errorf("could not do request: %w", errDoRequest)
		}
		defer getResponse.Body.Close()

		if getResponse.StatusCode >= 500 {
			return statusNotAvailable, nil
		}
		return statusAvailable, nil
	}
	if headResponse.StatusCode >= 500 {
		return statusNotAvailable, nil
	}
	return statusAvailable, nil
}

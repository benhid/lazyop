package main

import (
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	*http.Client

	// baseURL is the base URL of the OpenProject API including the version.
	baseURL string

	// username is the username used for authentication. It is always `apikey`.
	username string

	// apiKey is the API key used for authentication.
	apiKey string
}

func NewClient(baseURL, username, apiKey string) *Client {
	return &Client{
		Client:   http.DefaultClient,
		baseURL:  baseURL,
		username: username,
		apiKey:   apiKey,
	}
}

// doRequest performs an HTTP request with the given method, endpoint, and body.
func (c *Client) doRequest(method, endpoint string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.apiKey)

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", res.StatusCode, string(resBody))
	}

	return resBody, nil
}

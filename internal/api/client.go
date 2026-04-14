package api

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/saddledata/sd-cli/internal/config"
)

type Client struct {
	BaseUrl string
	ApiKey  string
	HTTP    *http.Client
}

func NewClient(ctx config.Context) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if ctx.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &Client{
		BaseUrl: ctx.ApiUrl,
		ApiKey:  ctx.ApiKey,
		HTTP: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) request(method, path string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/v1/cli%s", c.BaseUrl, path)
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Api-Key", c.ApiKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) Post(path string, body []byte) ([]byte, error) {
	return c.request("POST", path, body)
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.request("GET", path, nil)
}


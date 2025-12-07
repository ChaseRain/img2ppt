package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Options struct {
	Timeout    time.Duration
	MaxRetries int
}

type Client struct {
	client     *http.Client
	maxRetries int
}

func New(opts Options) *Client {
	return &Client{
		client: &http.Client{
			Timeout: opts.Timeout,
		},
		maxRetries: opts.MaxRetries,
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		req = req.WithContext(ctx)
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) Post(ctx context.Context, url string, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(ctx, req)
}

func (c *Client) PostJSON(ctx context.Context, url string, body []byte) (*http.Response, error) {
	return c.Post(ctx, url, "application/json", body)
}

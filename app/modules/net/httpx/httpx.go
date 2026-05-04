package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"mcop/app/utils/syncx"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const contentTypeJSON = "application/json"

// Client wraps the standard http.Client to provide additional functionality.
type Client struct {
	*http.Client
}

var (
	transport = otelhttp.NewTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	bbPool = syncx.NewPool(
		func() *bytes.Buffer {
			return bytes.NewBuffer(nil)
		},
	)
)

// Transport returns the default instrumented HTTP RoundTripper.
func Transport() http.RoundTripper {
	return transport
}

// NewClient returns a new http.Client with default timeout and instrumented transport.
func NewClient() *Client {
	return &Client{
		Client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

// NewRequest creates a new HTTP request with the provided context, method, URL, and body.
func NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}

// NewJSONRequest creates a new HTTP request with a JSON-encoded body and sets the Content-Type header.
func NewJSONRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	b := bbPool.Get()
	b.Reset()
	if err := json.NewEncoder(b).Encode(body); err != nil {
		return nil, err
	}
	req, err := NewRequest(ctx, method, url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	return req, nil
}

// DoJSON sends an HTTP request and decodes the JSON response body into dst if provided.
func (c *Client) DoJSON(req *http.Request, dst any) (*http.Response, error) {
	req.Header.Set("Accept", contentTypeJSON)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Type") != contentTypeJSON {
		return resp, ErrNotJSON
	}
	defer resp.Body.Close()
	if dst == nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

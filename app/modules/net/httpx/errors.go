// Package httpx provides HTTP utilities and error definitions for handling HTTP responses.
package httpx

import "fmt"

// ErrNotJSON is returned when the HTTP response body is not valid JSON.
var (
	ErrNotJSON = fmt.Errorf("response body is not valid JSON")
)

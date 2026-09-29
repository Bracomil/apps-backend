// internal/httpclient/options.go
package httpclient

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Option func(*Client)

func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.http = hc
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.http.Timeout = d
	}
}

func WithLimiter(l *rate.Limiter) Option {
	return func(c *Client) {
		c.limiter = l
	}
}

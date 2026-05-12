package main

import (
	"net/http"
	"time"
)

const (
	httpUserAgent = "gobump (https://github.com/lzap/gobump)"
	// httpClientTimeout bounds how long outbound HTTP (module proxy, GitHub API) may block.
	httpClientTimeout = 90 * time.Second
)

func newHTTPClient() *http.Client {
	return &http.Client{Timeout: httpClientTimeout}
}

func setDefaultHTTPHeaders(req *http.Request) {
	req.Header.Set("User-Agent", httpUserAgent)
}

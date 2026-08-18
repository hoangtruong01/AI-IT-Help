package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"eomp/packages/shared/pkg/errors"
)

// ReverseProxy maps a path prefix to a target downstream service URL
type ReverseProxy struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
	log    *slog.Logger
}

// NewReverseProxy creates a configured ReverseProxy instance
func NewReverseProxy(targetURL string, log *slog.Logger) (*ReverseProxy, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	// Custom error handler for proxy errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("gateway proxy forwarding error",
			slog.String("target", targetURL),
			slog.String("path", r.URL.Path),
			slog.Any("error", err),
		)
		errors.WriteHTTP(w, errors.New(http.StatusBadGateway, "BAD_GATEWAY", "downstream service unavailable"))
	}

	// Director modifying headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = parsedURL.Host
		// Ensure client IP / proto headers are preserved
		if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
			req.Header.Set("X-Forwarded-For", strings.Split(req.RemoteAddr, ":")[0])
		}
	}

	return &ReverseProxy{
		target: parsedURL,
		proxy:  proxy,
		log:    log,
	}, nil
}

// ServeHTTP delegates request handling to the reverse proxy
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}

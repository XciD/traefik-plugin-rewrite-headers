// nolint
package traefik_plugin_rewrite_headers

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
)

// Rewrite holds one rewrite body configuration.
type Rewrite struct {
	Header      string `json:"header,omitempty"`
	Regex       string `json:"regex,omitempty"`
	Replacement string `json:"replacement,omitempty"`
}

// Config holds the plugin configuration.
type Config struct {
	Rewrites []Rewrite `json:"rewrites,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewrite struct {
	header      string
	regex       *regexp.Regexp
	replacement string
}

type rewriteBody struct {
	name     string
	next     http.Handler
	rewrites []rewrite
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	rewrites := make([]rewrite, len(config.Rewrites))

	for i, rewriteConfig := range config.Rewrites {
		regex, err := regexp.Compile(rewriteConfig.Regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", rewriteConfig.Regex, err)
		}

		rewrites[i] = rewrite{
			header:      rewriteConfig.Header,
			regex:       regex,
			replacement: rewriteConfig.Replacement,
		}
	}

	return &rewriteBody{
		name:     name,
		next:     next,
		rewrites: rewrites,
	}, nil
}

func (r *rewriteBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrappedWriter := &responseWriter{
		writer:   rw,
		rewrites: r.rewrites,
	}

	r.next.ServeHTTP(wrappedWriter, req)
}

type responseWriter struct {
	writer   http.ResponseWriter
	rewrites []rewrite
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.writer.Write(bytes)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	for _, rewrite := range r.rewrites {
		headers := r.writer.Header().Values(rewrite.header)

		if len(headers) == 0 {
			continue
		}

		r.writer.Header().Del(rewrite.header)

		for _, header := range headers {
			value := rewrite.regex.ReplaceAllString(header, rewrite.replacement)
			r.writer.Header().Add(rewrite.header, value)
		}
	}

	r.writer.WriteHeader(statusCode)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.writer.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.writer)
	}

	return hijacker.Hijack()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

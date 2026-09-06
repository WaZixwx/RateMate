// Package proxy implements an HTTP/HTTPS forward proxy that injects a delay
// (computed by the limiter) before forwarding each request.
//
// HTTPS targets are handled via the CONNECT method: the proxy waits for the
// limiter, dials the target, replies 200 Connection Established, and then
// tunnels bytes in both directions until either side closes.
//
// Plain HTTP requests are forwarded with the proxy stripping hop-by-hop
// headers and rewriting the request URL to absolute form.
package proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WaZixwx/RateMate/internal/limiter"
	"github.com/WaZixwx/RateMate/internal/stats"
)

// Server is the delay-injecting forward proxy.
type Server struct {
	addr    string
	limiter *limiter.Limiter
	tracker *stats.Tracker
	logger  *log.Logger

	srv *http.Server

	running atomic.Bool
	stopCh  chan struct{}
	doneCh  chan struct{}
	mu      sync.Mutex
}

// New constructs a proxy listening on addr (e.g. ":8080").
func New(addr string, lim *limiter.Limiter, tr *stats.Tracker, logger *log.Logger) *Server {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	s := &Server{
		addr:    addr,
		limiter: lim,
		tracker: tr,
		logger:  logger,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	return s
}

// Running reports whether the proxy is currently accepting connections.
func (s *Server) Running() bool { return s.running.Load() }

// Start begins listening. It returns immediately; the server runs in a
// goroutine. Calling Start on an already-running server is a no-op.
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running.Load() {
		return nil
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.addr, err)
	}

	s.srv = &http.Server{
		Handler:           http.HandlerFunc(s.handle),
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       90 * time.Second,
		ErrorLog:          s.logger,
	}
	s.running.Store(true)
	// Reset stop channel in case of restart.
	s.stopCh = make(chan struct{})
	go func() {
		defer close(s.doneCh)
		_ = s.srv.Serve(ln)
	}()
	return nil
}

// Stop gracefully shuts the proxy down. It is safe to call on a stopped
// server.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running.Load() {
		return
	}
	s.running.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.srv.Shutdown(ctx)
}

// Addr returns the address the server is configured to listen on.
func (s *Server) Addr() string { return s.addr }

// SetAddr updates the listen address. The change takes effect on the next
// Start — the proxy does not support hot re-binding.
func (s *Server) SetAddr(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running.Load() {
		return
	}
	s.addr = addr
}

// handle dispatches between CONNECT (HTTPS) and normal HTTP forwarding.
func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	// Acquire the limiter before doing any work. The limiter records the
	// request immediately.
	startedAt := time.Now()
	waited, err := s.limiter.Wait(r.Context())
	if err != nil {
		http.Error(w, "limiter cancelled: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	host := r.Host
	if host == "" {
		host = r.URL.Host
	}

	if r.Method == http.MethodConnect {
		s.handleConnect(w, r, host, startedAt, waited)
		return
	}
	s.handleHTTP(w, r, host, startedAt, waited)
}

// handleConnect tunnels an HTTPS connection.
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request, host string, startedAt time.Time, waited time.Duration) {
	target := host
	if !strings.Contains(target, ":") {
		target += ":443"
	}

	dst, err := net.DialTimeout("tcp", target, 15*time.Second)
	if err != nil {
		s.tracker.Push(stats.Entry{
			Time: startedAt, Method: "CONNECT", Host: host,
			WaitedMs: waited.Milliseconds(), Err: err.Error(),
		})
		http.Error(w, "dial: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer dst.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack unsupported", http.StatusInternalServerError)
		return
	}
	src, _, err := hj.Hijack()
	if err != nil {
		s.logger.Printf("hijack: %v", err)
		return
	}
	defer src.Close()

	// Tell the client the tunnel is established.
	if _, err := fmt.Fprintf(src, "HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		return
	}

	bytesIn, bytesOut := tunnel(src, dst)

	s.tracker.Push(stats.Entry{
		Time: startedAt, Method: "CONNECT", Host: host,
		WaitedMs: waited.Milliseconds(), Status: 200,
		BytesIn: bytesIn, BytesOut: bytesOut,
	})
}

// handleHTTP forwards a plain HTTP request, stripping hop-by-hop headers.
func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request, host string, startedAt time.Time, waited time.Duration) {
	// Rebuild an outbound request. The proxy uses the absolute URL.
	outReq := r.Clone(r.Context())
	if r.URL.Host == "" {
		outReq.URL = &url.URL{Scheme: "http", Host: host, Path: r.URL.Path, RawQuery: r.URL.RawQuery}
	}
	removeHopHeaders(outReq.Header)

	tr := &http.Transport{}
	resp, err := tr.RoundTrip(outReq)
	if err != nil {
		s.tracker.Push(stats.Entry{
			Time: startedAt, Method: r.Method, Host: host, Path: r.URL.Path,
			WaitedMs: waited.Milliseconds(), Err: err.Error(),
		})
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	bytesOut, _ := io.Copy(w, resp.Body)

	s.tracker.Push(stats.Entry{
		Time: startedAt, Method: r.Method, Host: host, Path: r.URL.Path,
		WaitedMs: waited.Milliseconds(), Status: resp.StatusCode,
		BytesOut: bytesOut,
	})
}

// tunnel copies bytes both ways and returns (client→target, target→client) byte
// counts. Either side closing ends the tunnel.
func tunnel(client, target net.Conn) (in, out int64) {
	var mu sync.Mutex
	var bIn, bOut int64

	done := make(chan struct{}, 2)

	cp := func(dst io.Writer, src io.Reader, counter *int64) {
		n, _ := io.Copy(dst, src)
		mu.Lock()
		*counter = n
		mu.Unlock()
		// Closing the destination ends the other direction too.
		if c, ok := dst.(closeWriter); ok {
			_ = c.CloseWrite()
		} else if c, ok := dst.(io.Closer); ok {
			_ = c.Close()
		}
		done <- struct{}{}
	}

	go cp(target, client, &bIn)
	go cp(client, target, &bOut)

	<-done
	<-done

	return bIn, bOut
}

type closeWriter interface {
	io.Writer
	CloseWrite() error
}

// hop-by-hop headers (RFC 7230 §6.1) must not be forwarded.
var hopHeaders = []string{
	"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate",
	"Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

func removeHopHeaders(h http.Header) {
	for _, k := range hopHeaders {
		h.Del(k)
	}
}

// ParseAddr validates and normalises a listen address like ":8080".
func ParseAddr(port int) string {
	if port <= 0 || port > 65535 {
		port = 8080
	}
	return fmt.Sprintf(":%d", port)
}

var _ = bufio.NewReader // keep import for future use
var _ = errors.New

package middleware

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"time"
)

type Filter func(w WriterProxy, r *http.Request) bool

func AccessLogger(logger *slog.Logger, filters ...Filter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()
			lw := WrapWriter(w)

			defer func() {

				// Pass thru filters and skip early the code below, to prevent unnecessary processing.
				for _, filter := range filters {
					if !filter(lw, r) {
						return
					}
				}

				status := lw.Status()

				attributes := []slog.Attr{
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.Int("status_code", status),
					slog.Int("size_bytes", lw.BytesWritten()),
					slog.Duration("elapsed_ms", time.Since(start)),
					slog.String("remote_ip", r.RemoteAddr),
				}

				level := slog.LevelInfo
				if status >= http.StatusInternalServerError {
					level = slog.LevelError
				} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
					level = slog.LevelWarn
				}

				logger.LogAttrs(r.Context(), level, fmt.Sprintf("%d: %s", status, http.StatusText(status)), attributes...)
			}()
			next.ServeHTTP(lw, r)
		})
	}
}

func IgnorePath(urls ...string) Filter {
	return func(w WriterProxy, r *http.Request) bool {
		return !slices.Contains(urls, r.URL.Path)
	}
}

// WriterProxy is a proxy around an http.ResponseWriter that allows you to hook
// into various parts of the response process.
type WriterProxy interface {
	http.ResponseWriter
	// Status returns the HTTP status of the request, or 0 if one has not
	// yet been sent.
	Status() int
	// BytesWritten returns the total number of bytes sent to the client.
	BytesWritten() int
	// Unwrap returns the original proxied target.
	Unwrap() http.ResponseWriter
}

// WrapWriter wraps an http.ResponseWriter, returning a proxy that allows you to
// hook into various parts of the response process.
func WrapWriter(w http.ResponseWriter) WriterProxy {
	_, fl := w.(http.Flusher)
	_, hj := w.(http.Hijacker)
	_, rf := w.(io.ReaderFrom)

	bw := basicWriter{ResponseWriter: w}
	if fl && hj && rf {
		return &fancyWriter{bw}
	}
	if fl {
		return &flushWriter{bw}
	}
	return &bw
}

// basicWriter wraps a http.ResponseWriter that implements the minimal
// http.ResponseWriter interface.
type basicWriter struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
	bytes       int
}

func (b *basicWriter) WriteHeader(code int) {
	if !b.wroteHeader {
		b.code = code
		b.wroteHeader = true
		b.ResponseWriter.WriteHeader(code)
	}
}

func (b *basicWriter) Write(buf []byte) (int, error) {
	b.WriteHeader(http.StatusOK)
	n, err := b.ResponseWriter.Write(buf)
	b.bytes += n
	return n, err
}

func (b *basicWriter) maybeWriteHeader() {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
}

func (b *basicWriter) Status() int {
	return b.code
}

func (b *basicWriter) BytesWritten() int {
	return b.bytes
}

func (b *basicWriter) Unwrap() http.ResponseWriter {
	return b.ResponseWriter
}

// fancyWriter is a writer that additionally satisfies http.Flusher,
// http.Hijacker, and io.ReaderFrom. It exists for the common case
// of wrapping the http.ResponseWriter that package http gives you, in order to
// make the proxied object support the full method set of the proxied object.
type fancyWriter struct {
	basicWriter
}

func (f *fancyWriter) Flush() {
	fl := f.basicWriter.ResponseWriter.(http.Flusher) // Panics if not Flusher
	fl.Flush()
}

func (f *fancyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.basicWriter.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func (f *fancyWriter) ReadFrom(r io.Reader) (int64, error) {
	rf := f.basicWriter.ResponseWriter.(io.ReaderFrom)
	f.basicWriter.maybeWriteHeader()

	n, err := rf.ReadFrom(r)
	f.bytes += int(n)
	return n, err
}

type flushWriter struct {
	basicWriter
}

func (f *flushWriter) Flush() {
	fl := f.basicWriter.ResponseWriter.(http.Flusher)
	fl.Flush()
}

var (
	_ http.Flusher  = &fancyWriter{}
	_ http.Hijacker = &fancyWriter{}
	_ io.ReaderFrom = &fancyWriter{}
	_ http.Flusher  = &flushWriter{}
)

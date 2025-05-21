package api

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type responseCodeWriter struct {
	http.ResponseWriter
	statusCode int
	hijacked   bool
}

func newResponseWriter(w http.ResponseWriter) *responseCodeWriter {
	return &responseCodeWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseCodeWriter) WriteHeader(code int) {
	if rw.hijacked {
		return
	}
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseCodeWriter) Write(b []byte) (int, error) {
	if rw.hijacked {
		return 0, http.ErrHijacked
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *responseCodeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("http.Hijacker not supported")
	}
	conn, bufrw, err := h.Hijack()
	if err == nil {
		rw.hijacked = true
		rw.statusCode = http.StatusSwitchingProtocols
	}
	return conn, bufrw, err
}

func (rw *responseCodeWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

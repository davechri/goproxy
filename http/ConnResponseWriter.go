package http

import (
	"log"
	"net"
	"net/http"
	"strconv"
)

// net.Conn ResponseWriter
type connResponseWriter struct {
	conn           net.Conn
	header         http.Header
	headersWritten *bool
}

func (w connResponseWriter) Header() http.Header {
	log.Println("connResponseWriter Header()", w.header)
	return w.header
}

func (w connResponseWriter) Write(data []byte) (int, error) {
	log.Printf("connResponseWriter Write() %d\n", len(data))
	if !*w.headersWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.conn.Write(data)
}

func (w connResponseWriter) WriteHeader(statusCode int) {
	log.Printf("connResponseWriter WriteHeader(%d) %v", statusCode, w.header)
	w.conn.Write([]byte("HTTP/1.1 " + strconv.Itoa(statusCode) + "\r\n"))
	w.header.Write(w.conn)
	w.conn.Write([]byte("\r\n"))
	*w.headersWritten = true
}

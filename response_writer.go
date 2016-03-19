package lion

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
)

// ResponseWriter is the proxy responseWriter
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	Status() int
	BytesWritten() int
	Tee(io.Writer)
	Unwrap() http.ResponseWriter
}

// WrapResponseWriter wraps an http.ResponseWriter and returns a ResponseWriter
func WrapResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &basicWriter{ResponseWriter: w}
}

var _ ResponseWriter = (*basicWriter)(nil)
var _ http.ResponseWriter = (*basicWriter)(nil)

type basicWriter struct {
	http.ResponseWriter
	code  int
	bytes int
	tee   io.Writer
}

func (b *basicWriter) Header() http.Header {
	return b.ResponseWriter.Header()
}

func (b *basicWriter) WriteHeader(code int) {
	if !b.Written() {
		b.ResponseWriter.WriteHeader(code)
		b.code = code
	}
}

func (b *basicWriter) Write(data []byte) (int, error) {
	if !b.Written() {
		b.WriteHeader(http.StatusOK)
	}
	size, err := b.ResponseWriter.Write(data)
	b.bytes += size
	return size, err
}

func (b *basicWriter) Written() bool {
	return b.Status() != 0
}

func (b *basicWriter) BytesWritten() int {
	return b.bytes
}

func (b *basicWriter) Status() int {
	return b.code
}

func (b *basicWriter) Tee(w io.Writer) {
	b.tee = w
}

func (b *basicWriter) Unwrap() http.ResponseWriter {
	return b.ResponseWriter
}

func (b *basicWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := b.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (b *basicWriter) CloseNotify() <-chan bool {
	return b.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (b *basicWriter) Flush() {
	fl, ok := b.ResponseWriter.(http.Flusher)
	if ok {
		fl.Flush()
	}
}

func (b *basicWriter) ReadFrom(r io.Reader) (int64, error) {
	if b.tee != nil {
		return io.Copy(b, r)
	}
	rf := b.ResponseWriter.(io.ReaderFrom)
	if !b.Written() {
		b.ResponseWriter.WriteHeader(http.StatusOK)
	}
	return rf.ReadFrom(r)
}

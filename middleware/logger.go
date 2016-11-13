package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/celrenheit/lion"
	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var magenta = color.New(color.FgHiMagenta).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var hiBlue = color.New(color.FgHiBlue).SprintFunc()

// Logger is a middlewares that logs incoming http requests
type Logger struct {
	*log.Logger
}

// NewLogger creates a new Logger
func NewLogger() lion.Middleware {
	return &Logger{
		Logger: lionLogger,
	}
}

// ServeNext implements the Middleware interface for Logger.
// It wraps the corresponding http.ResponseWriter and saves statistics about the status code returned, the number of bytes written and the time that requests took.
func (l *Logger) ServeNext(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		res := wrapResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(res, r)

		l.Printf("%s %s | %s | %dB in %v from %s", magenta(r.Method), hiBlue(r.URL.Path), statusColor(res.Status()), res.BytesWritten(), timeColor(time.Since(start)), r.RemoteAddr)
	})
}

func statusColor(status int) string {
	msg := fmt.Sprintf("%d %s", status, http.StatusText(status))
	switch {
	case status < 200:
		return blue(msg)
	case status < 300:
		return green(msg)
	case status < 400:
		return cyan(msg)
	case status < 500:
		return yellow(msg)
	default:
		return red(msg)
	}
}

func timeColor(dur time.Duration) string {
	switch {
	case dur < 500*time.Millisecond:
		return green(dur)
	case dur < 5*time.Second:
		return yellow(dur)
	default:
		return red(dur)
	}
}

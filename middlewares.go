package lion

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"

	"golang.org/x/net/context"
)

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).SprintFunc()
var magenta = color.New(color.FgHiMagenta).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var cyan = color.New(color.FgCyan).SprintFunc()
var hiBlue = color.New(color.FgHiBlue).SprintFunc()
var lionColor = color.New(color.Italic, color.FgHiGreen).SprintFunc()

// Classic creates a new router instance with default middlewares: Recovery, RealIP, Logger and Static.
// The static middleware instance is initiated with a directory named "public" located relatively to the current working directory.
func Classic() *Router {
	return New(NewRecovery(), RealIP(), NewLogger(), NewStatic(http.Dir("public")))
}

var lionLogger = log.New(os.Stdout, lionColor("[lion]")+" ", log.Ldate|log.Ltime)

// Logger is a middlewares that logs incoming http requests
type Logger struct {
	*log.Logger
}

// NewLogger creates a new Logger
func NewLogger() *Logger {
	return &Logger{
		Logger: lionLogger,
	}
}

// ServeNext implements the Middleware interface for Logger.
// It wraps the corresponding http.ResponseWriter and saves statistics about the status code returned, the number of bytes written and the time that requests took.
func (l *Logger) ServeNext(next Handler) Handler {

	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {

		res := WrapResponseWriter(w)
		start := time.Now()

		next.ServeHTTPC(c, res, r)

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

// Recovery is a middleware that recovers from panics
// Taken from https://github.com/codegangsta/negroni/blob/master/recovery.go
type Recovery struct {
	Logger     *log.Logger
	PrintStack bool
	StackAll   bool
	StackSize  int
}

// NewRecovery creates a new Recovery instance
func NewRecovery() *Recovery {
	return &Recovery{
		Logger:     lionLogger,
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

// ServeNext is the method responsible for recovering from a panic
func (rec *Recovery) ServeNext(next Handler) Handler {
	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				stack := make([]byte, rec.StackSize)
				stack = stack[:runtime.Stack(stack, rec.StackAll)]

				f := "PANIC: %s\n%s"
				rec.Logger.Printf(f, err, stack)

				if rec.PrintStack {
					fmt.Fprintf(w, f, err, stack)
				}

			}
		}()

		next.ServeHTTPC(c, w, r)
	})
}

// Static is a middleware handler that serves static files in the given directory/filesystem.
// Taken from https://github.com/codegangsta/negroni/blob/master/static.go
type Static struct {
	// Dir is the directory to serve static files from
	Dir http.FileSystem
	// Prefix is the optional prefix used to serve the static directory content
	Prefix string
	// IndexFile defines which file to serve as index if it exists.
	IndexFile string
}

// NewStatic returns a new instance of Static
func NewStatic(directory http.FileSystem) *Static {
	return &Static{
		Dir:       directory,
		Prefix:    "",
		IndexFile: "index.html",
	}
}

// ServeNext tries to find a file in the directory
func (s *Static) ServeNext(next Handler) Handler {
	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" && r.Method != "HEAD" {
			next.ServeHTTPC(c, w, r)
			return
		}

		file := r.URL.Path
		// if we have a prefix, filter requests by stripping the prefix
		if s.Prefix != "" {
			if !strings.HasPrefix(file, s.Prefix) {
				next.ServeHTTPC(c, w, r)
				return
			}
			file = file[len(s.Prefix):]
			if file != "" && file[0] != '/' {
				next.ServeHTTPC(c, w, r)
				return
			}
		}
		f, err := s.Dir.Open(file)
		if err != nil {
			// discard the error?
			next.ServeHTTPC(c, w, r)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			next.ServeHTTPC(c, w, r)
			return
		}

		// try to serve index file
		if fi.IsDir() {
			// redirect if missing trailing slash
			if !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
				return
			}

			file = path.Join(file, s.IndexFile)
			f, err = s.Dir.Open(file)
			if err != nil {
				next.ServeHTTPC(c, w, r)
				return
			}
			defer f.Close()

			fi, err = f.Stat()
			if err != nil || fi.IsDir() {
				next.ServeHTTPC(c, w, r)
				return
			}
		}

		http.ServeContent(w, r, file, fi.ModTime(), f)
	})
}

// MaxAge is a middleware that defines the max duration headers
func MaxAge(dur time.Duration) Middleware {
	return MaxAgeWithFilter(dur, func(c context.Context, w http.ResponseWriter, r *http.Request) bool { return true })
}

// MaxAgeWithFilter is a middleware that defines the max duration headers with a filter function.
// If the filter returns true then the headers will be set. Otherwise, if it returns false the headers will not be set.
func MaxAgeWithFilter(dur time.Duration, filter func(c context.Context, w http.ResponseWriter, r *http.Request) bool) Middleware {
	if filter == nil {
		filter = func(c context.Context, w http.ResponseWriter, r *http.Request) bool { return false }
	}
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			if !filter(c, w, r) {
				return
			}
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", int(dur.Seconds())))
			next.ServeHTTPC(c, w, r)
		})
	})
}

// NoCache middleware sets headers to disable browser caching.
// Inspired by https://github.com/mytrile/nocache
func NoCache() Middleware {
	var epoch = time.Unix(0, 0).Format(time.RFC1123)

	return noCache{
		responseHeaders: map[string]string{
			"Expires":         epoch,
			"Cache-Control":   "no-cache, private, must-revalidate, max-age=0",
			"Pragma":          "no-cache",
			"X-Accel-Expires": "0",
		},
		etagHeaders: []string{
			"ETag",
			"If-Modified-Since",
			"If-Match",
			"If-Note-Match",
			"If-Range",
			"If-Unmodified-Since",
		},
	}
}

type noCache struct {
	responseHeaders map[string]string
	etagHeaders     []string
}

func (n noCache) ServeNext(next Handler) Handler {
	return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		// Delete ETag headers
		for _, v := range n.etagHeaders {
			if r.Header.Get(v) != "" {
				r.Header.Del(v)
			}
		}

		// Add nocache headers
		for k, v := range n.responseHeaders {
			w.Header().Set(k, v)
		}

	})
}

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results
// of parsing either the X-Forwarded-For header or the X-Real-IP header (in that
// order).
//
// This middleware should be inserted fairly early in the middleware stack to
// ensure that subsequent layers (e.g., request loggers) which examine the
// RemoteAddr will see the intended value.
//
// You should only use this middleware if you can trust the headers passed to
// you (in particular, the two headers this middleware uses), for example
// because you have placed a reverse proxy like HAProxy or nginx in front of
// Goji. If your reverse proxies are configured to pass along arbitrary header
// values from the client, or if you use this middleware without a reverse
// proxy, malicious clients will be able to make you very sad (or, depending on
// how you're using RemoteAddr, vulnerable to an attack of some sort).
// Taken from https://github.com/zenazn/goji/blob/master/web/middleware/realip.go
func RealIP() Middleware {
	return MiddlewareFunc(func(next Handler) Handler {
		return HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
			if rip := realIP(r); rip != "" {
				r.RemoteAddr = rip
			}
			next.ServeHTTPC(c, w, r)
		})
	})
}

func realIP(r *http.Request) string {
	var ip string

	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	}

	return ip
}

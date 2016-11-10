package middleware

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/celrenheit/lion"
	"github.com/fatih/color"
)

var lionColor = color.New(color.Italic, color.FgHiGreen).SprintFunc()
var lionLogger = log.New(os.Stdout, lionColor("[lion]")+" ", log.Ldate|log.Ltime)

func Basic() lion.Middlewares {
	return lion.Middlewares{NewRecovery(), NewRealIP()}
}

// Classic creates a new router instance with default middlewares: Recovery, RealIP, Logger.
// The static middleware instance is initiated with a directory named "public" located relatively to the current working directory.
func Classic() lion.Middlewares {
	return lion.Middlewares{NewRecovery(), NewRealIP(), NewLogger()}
}

// NewLogger creates a new Logger
func NewLogger() lion.Middleware {
	return &Logger{
		Logger: lionLogger,
	}
}

// NewRecovery creates a new Recovery instance
func NewRecovery() lion.Middleware {
	return &Recovery{
		Logger:     lionLogger,
		PrintStack: false,
		StackAll:   false,
		StackSize:  1024 * 8,
	}
}

// NewStatic returns a new instance of Static
func NewStatic(directory http.FileSystem) lion.Middleware {
	return &Static{
		Dir:       directory,
		Prefix:    "",
		IndexFile: "index.html",
	}
}

// NoCache middleware sets headers to disable browser caching.
// Inspired by https://github.com/mytrile/nocache
func NewNoCache() lion.Middleware {
	var epoch = time.Unix(0, 0).Format(time.RFC1123)

	return NoCache{
		ResponseHeaders: map[string]string{
			"Expires":         epoch,
			"Cache-Control":   "no-cache, private, must-revalidate, max-age=0",
			"Pragma":          "no-cache",
			"X-Accel-Expires": "0",
		},
		EtagHeaders: []string{
			"ETag",
			"If-Modified-Since",
			"If-Match",
			"If-Note-Match",
			"If-Range",
			"If-Unmodified-Since",
		},
	}
}

func NewRealIP() lion.Middleware {
	return RealIP{}
}

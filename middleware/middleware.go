package middleware

import (
	"log"
	"os"

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

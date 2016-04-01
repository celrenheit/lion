package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

type todoList struct{}

func (t todoList) Uses() lion.Middlewares {
	return lion.Middlewares{lion.NewLogger()}
}

func (t todoList) GetMiddlewares() lion.Middlewares {
	return lion.Middlewares{lion.NewRecovery()}
}

func (t todoList) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
	// Should be catched by GetMiddlewares()'s Recovery middleware
	panic("test")
}

func main() {
	l := lion.New()
	l.Resource("/todos", todoList{})
	l.Run()
}

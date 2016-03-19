package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

type TodoList struct{}

func (t TodoList) Uses() lion.Middlewares {
	return lion.Middlewares{lion.NewLogger()}
}

func (t TodoList) GetMiddlewares() lion.Middlewares {
	return lion.Middlewares{lion.NewRecovery()}
}

func (t TodoList) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
	// Should be catched by GetMiddlewares()'s Recovery middleware
	panic("test")
}

func main() {
	l := lion.New()
	l.Resource("/todos", TodoList{})
	l.Run()
}

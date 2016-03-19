package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

// Open your web browser at http://localhost:3000/api/v1/todos

type api struct{}

func (t api) Base() string { return "/api" }

func (t api) Routes(r *lion.Router) {
	r.Module(v1{})
}

// Attach Get methods to a Module.
// ====> A Module is also a Resource.
func (t api) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
    Description of available apis
    Go to: http://localhost:3000/api/v1/todos
    `)
}

type v1 struct{}

func (t v1) Base() string { return "/v1" }

func (t v1) Routes(r *lion.Router) {
	r.Resource("/todos", todoList{})
}

type todoList struct{}

func (t todoList) Uses() lion.Middlewares {
	return lion.Middlewares{lion.NewLogger()}
}

func (t todoList) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
}

func main() {
	l := lion.New()
	l.Module(api{})
	l.Run()
}

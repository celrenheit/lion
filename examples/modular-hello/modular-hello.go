package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

func main() {
	l := lion.Classic()
	api := l.Group("/api")
	api.Module(Products{})
	l.Run()
}

// Products module is accessible at url: /api/products
// It handles getting a list of products or creating a new product
type Products struct{}

func (p Products) Base() string {
	return "/products"
}

func (p Products) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Fetching all products")
}

func (p Products) Post(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Creating a new product")
}

func (p Products) Routes(r *lion.Router) {
	// Defining a resource for getting, editing and deleting a single product
	r.Resource("/:id", OneProduct{})
}

// OneProduct resource is accessible at url: /api/products/:id
// It handles getting, editing and deleting a single product
type OneProduct struct{}

func (p OneProduct) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	id := lion.Param(c, "id")
	fmt.Fprintf(w, "Getting product: %s", id)
}

func (p OneProduct) Put(c context.Context, w http.ResponseWriter, r *http.Request) {
	id := lion.Param(c, "id")
	fmt.Fprintf(w, "Updating article: %s", id)
}

func (p OneProduct) Delete(c context.Context, w http.ResponseWriter, r *http.Request) {
	id := lion.Param(c, "id")
	fmt.Fprintf(w, "Deleting article: %s", id)
}

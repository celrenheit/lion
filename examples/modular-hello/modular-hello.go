package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
)

func main() {
	l := lion.Classic()
	api := l.Group("/api")
	api.Module(products{})
	l.Run()
}

// products module is accessible at url: /api/products
// It handles getting a list of products or creating a new product
type products struct{}

func (p products) Base() string {
	return "/products"
}

func (p products) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Fetching all products")
}

func (p products) Post(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Creating a new product")
}

func (p products) Routes(r *lion.Router) {
	// Defining a resource for getting, editing and deleting a single product
	r.Resource("/:id", oneProduct{})
}

// oneProduct resource is accessible at url: /api/products/:id
// It handles getting, editing and deleting a single product
type oneProduct struct{}

func (p oneProduct) Get(w http.ResponseWriter, r *http.Request) {
	id := lion.Param(r, "id")
	fmt.Fprintf(w, "Getting product: %s", id)
}

func (p oneProduct) Put(w http.ResponseWriter, r *http.Request) {
	id := lion.Param(r, "id")
	fmt.Fprintf(w, "Updating article: %s", id)
}

func (p oneProduct) Delete(w http.ResponseWriter, r *http.Request) {
	id := lion.Param(r, "id")
	fmt.Fprintf(w, "Deleting article: %s", id)
}

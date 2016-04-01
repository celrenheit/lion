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
	api.Module(products{})
	l.Run()
}

// products module is accessible at url: /api/products
// It handles getting a list of products or creating a new product
type products struct{}

func (p products) Base() string {
	return "/products"
}

func (p products) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Fetching all products")
}

func (p products) Post(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Creating a new product")
}

func (p products) Routes(r *lion.Router) {
	// Defining a resource for getting, editing and deleting a single product
	r.Resource("/:id", oneProduct{})
}

// oneProduct resource is accessible at url: /api/products/:id
// It handles getting, editing and deleting a single product
type oneProduct struct{}

func (p oneProduct) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	id := lion.Param(c, "id")
	fmt.Fprintf(w, "Getting product: %s", id)
}

func (p oneProduct) Put(c context.Context, w http.ResponseWriter, r *http.Request) {
	id := lion.Param(c, "id")
	fmt.Fprintf(w, "Updating article: %s", id)
}

func (p oneProduct) Delete(c context.Context, w http.ResponseWriter, r *http.Request) {
	id := lion.Param(c, "id")
	fmt.Fprintf(w, "Deleting article: %s", id)
}

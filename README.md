# Lion [![Build Status](https://img.shields.io/travis/celrenheit/lion.svg?style=flat-square)](https://travis-ci.org/celrenheit/lion) [![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/celrenheit/lion) [![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/celrenheit/lion?style=flat-square)](https://goreportcard.com/report/github.com/celrenheit/lion)

Lion is a [fast](#benchmarks) HTTP router for Go with support for middlewares for building modern scalable modular REST APIs.

![Lion's Hello World GIF](https://raw.githubusercontent.com/celrenheit/gifs/master/lion/hello-speed2-sm.min.gif)

## Lion v2

If you are starting a new project, please consider starting out using the [_v2_](https://github.com/celrenheit/lion/tree/v2) 
and the new [documentation](https://godoc.org/gopkg.in/celrenheit/lion.v2) branch. It contains a few breaking changes.

The most important one is that it now uses native [http.Handler](https://golang.org/pkg/net/http/#Handler).

# Important

If you are using lion v1, please change your import path to `gopkg.in/celrenheit/lion.v1`.

You can get lion via:
```shell
go get -u gopkg.in/celrenheit/lion.v1
```

## Features

* **Go1-like guarantee**: The API will **not** change in Lion v2.x (once released).
* **Modular**: You can define your own modules to easily build a scalable architecture
* **RESTful**: You can define modules to groups http resources together.
* **Subdomains**: Select which subdomains, hosts a router can match. You can specify it a param or a wildcard e.g. `*.example.org`. More infos [here](#match-hosts).
* **Near zero garbage**: Lion generates near zero garbage[*](#benchmarks). The allocations generated comes from the net/http.Request.WithContext() works.
It makes a shallow copy of the request.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
## Table of contents

- [Install/Update](#installupdate)
- [Hello World](#hello-world)
- [Getting started with modules and resources](#getting-started-with-modules-and-resources)
  - [Using net/http.Handler](#using-nethttphandler)
  - [Using net/http.HandlerFunc](#using-nethttphandlerfunc)
- [Middlewares](#middlewares)
  - [Using Named Middlewares](#using-named-middlewares)
  - [Using Third-Party Middlewares](#using-third-party-middlewares)
    - [Negroni](#negroni)
- [Matching Subdomains/Hosts](#matching-subdomainshosts)
- [Resources](#resources)
- [Examples](#examples)
  - [Using GET, POST, PUT, DELETE http methods](#using-get-post-put-delete-http-methods)
  - [Using middlewares](#using-middlewares)
  - [Group routes by a base path](#group-routes-by-a-base-path)
  - [Mounting a router into a base path](#mounting-a-router-into-a-base-path)
  - [Default middlewares](#default-middlewares)
  - [Custom Middlewares](#custom-middlewares)
    - [Custom Logger example](#custom-logger-example)
- [Benchmarks](#benchmarks)
- [Contributing](#contributing)
- [License](#license)
- [Todo](#todo)
- [Credits](#credits)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Install/Update

Lion requires Go 1.7+:

```shell
$ go get -u gopkg.in/celrenheit/lion.v1
```

Next versions of Lion will support the latest Go version and the previous one. 
For example, when Go 1.8 is out, Lion will support Go 1.7 and Go 1.8.

## Hello World

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"context"
)

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func Hello(w http.ResponseWriter, r *http.Request) {
	name := lion.Param(r, "name")
	fmt.Fprintf(w, "Hello %s!",name)
}

func main() {
	l := lion.Classic()
	l.GetFunc("/", Home)
	l.GetFunc("/hello/:name", Hello)
	l.Run()
}
```

Try it yourself by running the following command from the current directory:

```shell
$ go run examples/hello/hello.go
```

## Getting started with modules and resources

We are going to build a sample products listing REST api (without database handling to keep it simple):

```go

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

func (p Products) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Fetching all products")
}

func (p Products) Post(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Creating a new product")
}

func (p Products) Routes(r *lion.Router) {
	// Defining a resource for getting, editing and deleting a single product
	r.Resource("/:id", OneProduct{})
}

// OneProduct resource is accessible at url: /api/products/:id
// It handles getting, editing and deleting a single product
type OneProduct struct{}

func (p OneProduct) Get(w http.ResponseWriter, r *http.Request) {
	id := lion.Param(r, "id")
	fmt.Fprintf(w, "Getting product: %s", id)
}

func (p OneProduct) Put(w http.ResponseWriter, r *http.Request) {
	id := lion.Param(r, "id")
	fmt.Fprintf(w, "Updating article: %s", id)
}

func (p OneProduct) Delete(w http.ResponseWriter, r *http.Request) {
	id := lion.Param(r, "id")
	fmt.Fprintf(w, "Deleting article: %s", id)
}
```

Try it yourself. Run:
```shell
$ go run examples/modular-hello/modular-hello.go
```

Open your web browser to [http://localhost:3000/api/products](http://localhost:3000/api/products) or [http://localhost:3000/api/products/123](http://localhost:3000/api/products/123). You should see "_Fetching all products_" or "_Getting product: 123_".


### Using net/http.Handler

Handlers should implement the native net/http.Handler:

```go
l.Get("/get", get)
l.Post("/post", post)
l.Put("/put", put)
l.Delete("/delete", delete)
```

### Using net/http.HandlerFunc

You can use native net/http.Handler (`func(w http.ResponseWriter, r *http.Request)`):

```go
func myHandlerFunc(w http.ResponseWriter, r *http.Request)  {
  fmt.Fprintf(w, "Hi!")
}

l.GetFunc("/get", myHandlerFunc)
l.PostFunc("/post", myHandlerFunc)
l.PutFunc("/put", myHandlerFunc)
l.DeleteFunc("/delete", myHandlerFunc)
```

## Middlewares

Middlewares should implement the Middleware interface:

```go
type Middleware interface {
	ServeNext(http.Handler) http.Handler
}
```

The ServeNext function accepts a http.Handler and returns a http.Handler.

You can also use MiddlewareFuncs which are basically just `func(http.Handler) http.Handler` 

For example:

```go
func middlewareFunc(next Handler) Handler  {
	return next
}
```

### Using Named Middlewares

Named middlewares are designed to be able to reuse a previously defined middleware. For example, if you have a EnsureAuthenticated middleware that check whether a user is logged in.
You can define it once and reuse later in your application.

```go
l := lion.New()
l.Define("EnsureAuthenticated", NewEnsureAuthenticatedMiddleware())
```

To reuse it later in your application, you can use the `UseNamed` method. 
If it cannot find the named middleware if the current Router instance it will try to find it in the parent router.
If a named middleware is not found it will panic.

```go
api := l.Group("/api")
api.UseNamed("EnsureAuthenticated")
```

### Using Third-Party Middlewares

#### Negroni 

In v1, negroni was supported directly using UseNegroni.
It still works but you will have to use .UseNext and pass it a [negroni.HandlerFunc](https://godoc.org/github.com/urfave/negroni#HandlerFunc): 
`func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)`

This way if you prefer to use this kind of middleware, you can.

You can use [Negroni](https://github.com/codegangsta/negroni) middlewares
you can find a list of third party middlewares [here](https://github.com/codegangsta/negroni#third-party-middleware)

```go
l := lion.New()
l.UseNext(negroni.NewRecovery().ServeHTTP)
l.Run()
```

## Matching Subdomains/Hosts

You can match a specific or multiple hosts. You can use patterns in the same way they are currently used for routes with only some [edge cases](https://godoc.org/github.com/celrenheit/lion#Router.Host).
The main difference is that you will have to use the '**$**' character instead of '**:**' to define a parameter.


admin.example.com			will match			admin.example.com
$username.blog.com			will match			messi.blog.com
					will not match			my.awesome.blog.com
*.example.com				will match			my.admin.example.com

```go
l := New()

// Group by /api basepath
api := l.Group("/api")

// Specific to v1
v1 := api.Subrouter().
	Host("v1.example.org")

v1.Get("/", v1Handler)

// Specific to v2
v2 := api.Subrouter().
	Host("v2.example.org")

v2.Get("/", v2Handler)
l.Run()
```

## Resources

You can define a resource to represent a REST, CRUD api resource.
You define global middlewares using Uses() method. For defining custom middlewares for each http method, you have to create a function which name is composed of the http method suffixed by "Middlewares". For example, if you want to define middlewares for the Get method you will have to create a method called: **GetMiddlewares()**.

A resource is defined by the following methods. **Everything is optional**:
```go

// Global middlewares for the resource (Optional)
Uses() Middlewares

// Middlewares for the http methods (Optional)
GetMiddlewares() Middlewares
PostMiddlewares() Middlewares
PutMiddlewares() Middlewares
DeleteMiddlewares() Middlewares


// HandlerFuncs for each HTTP Methods (Optional)
Get(w http.ResponseWriter, r *http.Request)
Post(w http.ResponseWriter, r *http.Request)
Put(w http.ResponseWriter, r *http.Request)
Delete(w http.ResponseWriter, r *http.Request)
```

**_Example_**:

```go
package main

type todolist struct{}

func (t todolist) Uses() lion.Middlewares {
	return lion.Middlewares{lion.NewLogger()}
}

func (t todolist) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "getting todos")
}

func main() {
	l := lion.New()
	l.Resource("/todos", todolist{})
	l.Run()
}
```


## Modules

Modules are a way to modularize an api which can then define submodules, subresources and custom routes.
A module is defined by the following methods:

```go
// Required: Base url pattern of the module
Base() string

// Routes accepts a Router instance. This method is used to define the routes of this module.
// Each routes defined are relative to the Base() url pattern
Routes(*Router)

// Optional: Requires named middlewares. Refer to Named Middlewares section
Requires() []string
```

```go
package main

type api struct{}

// Required: Base url
func (t api) Base() string { return "/api" }

// Required: Here you can declare sub-resources, submodules and custom routes.
func (t api) Routes(r *lion.Router) {
	r.Module(v1{})
	r.Get("/custom", t.CustomRoute)
}

// Optional: Attach Get method to this Module.
// ====> A Module is also a Resource.
func (t api) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This also a resource accessible at http://localhost:3000/api")
}

// Optional: Defining custom routes
func (t api) CustomRoute(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This a custom route for this module http://localhost:3000/api/")
}

func main() {
	l := lion.New()
	// Registering the module
	l.Module(api{})
	l.Run()
}
```

## Examples

### Using GET, POST, PUT, DELETE http methods

```go
l := lion.Classic()

// Using Handlers
l.Get("/get", get)
l.Post("/post", post)
l.Put("/put", put)
l.Delete("/delete", delete)

// Using functions
l.GetFunc("/get", getFunc)
l.PostFunc("/post", postFunc)
l.PutFunc("/put", putFunc)
l.DeleteFunc("/delete", deleteFunc)

l.Run()
```

### Using middlewares

```go
func main() {
	l := lion.Classic()

	// Using middleware
	l.Use(lion.NewLogger())

	// Using middleware functions
	l.UseFunc(someMiddlewareFunc)

	l.GetFunc("/hello/:name", Hello)

	l.Run()
}
```


### Group routes by a base path

```go
l := lion.Classic()
api := l.Group("/api")

v1 := l.Group("/v1")
v1.GetFunc("/somepath", gettingFromV1)

v2 := l.Group("/v2")
v2.GetFunc("/somepath", gettingFromV2)

l.Run()
```

### Mounting a router into a base path


```go
l := lion.Classic()

sub := lion.New()
sub.GetFunc("/somepath", getting)


l.Mount("/api", sub)
```

### Default middlewares

`lion.Classic()` creates a router with default middlewares (Recovery, RealIP, Logger, Static).
If you wish to create a blank router without any middlewares you can use `lion.New()`.

```go
func main()  {
	// This a no middlewares registered
	l := lion.New()
	l.Use(lion.NewLogger())

	l.GetFunc("/hello/:name", Hello)

	l.Run()
}
```

### Custom Middlewares

Custom middlewares should implement the Middleware interface:

```go
type Middleware interface {
	ServeNext(Handler) Handler
}
```

You can also make MiddlewareFuncs to use using `.UseFunc()` method.
It has to accept a Handler and return a Handler:
```go
func(next Handler) Handler
```


#### Custom Logger example

```go
type logger struct{}

func (*logger) ServeNext(next lion.Handler) lion.Handler {
	return lion.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTPC(c, w, r)

		fmt.Printf("Served %s in %s\n", r.URL.Path, time.Since(start))
	})
}
```

Then in the main function you can use the middleware using:

```go
l := lion.New()

l.Use(&logger{})
l.GetFunc("/hello/:name", Hello)
l.Run()
```

## Benchmarks

TODO: Update this when v2 is released.

## Contributing

Want to contribute to Lion ? Awesome! Feel free to submit an issue or a pull request.

Here are some ways you can help:
* Report bugs
* Share a middleware or a module
* Improve code/documentation
* Propose new features
* and more...

## License

https://github.com/celrenheit/lion/blob/master/LICENSE

## Todo

* [x] Support for Go 1.7 context
* [x] Host matching
* [x] Automatic OPTIONS handler
* [ ] Modules
  * [ ] JWT Auth module
* [x] Better static file handling
* [ ] More documentation

## Credits

Lion v1 was inspired by [pressly/chi](https://github.com/pressly/chi). 
If Lion is not the right http router for you, check out chi.

Parts of Lion taken for other projects:
* [Negroni](https://github.com/codegangsta/negroni)
	* Static and Recovery middlewares are taken from Negroni
* [Goji](https://github.com/zenazn/goji/)
	* RealIP middleware is taken from goji
* [Gin](https://github.com/gin-gonic/gin)
	* ResponseWriter and Logger inspiration are inspired by gin
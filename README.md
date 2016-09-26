# Lion [![Build Status](https://img.shields.io/travis/celrenheit/lion.svg?style=flat-square)](https://travis-ci.org/celrenheit/lion) [![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/celrenheit/lion) [![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/celrenheit/lion?style=flat-square)](https://goreportcard.com/report/github.com/celrenheit/lion)

Lion is a [fast](#benchmarks) HTTP router for Go with support for middlewares for building modern scalable modular REST APIs.

![Lion's Hello World GIF](https://raw.githubusercontent.com/celrenheit/gifs/master/lion/hello-speed2-sm.min.gif)

## Features

* **Context-Aware**: Lion uses the de-facto standard [net/Context](https://golang.org/x/net/context) for storing route params and sharing variables between middlewares and HTTP handlers. It [_could_](https://github.com/golang/go/issues/14660) be integrated in the [standard library](https://github.com/golang/go/issues/13021) for Go 1.7 in 2016.
* **Modular**: You can define your own modules to easily build a scalable architecture
* **REST friendly**: You can define modules to groups http resources together.
* **Host**: Match hosts. Each host can get its own content.
* **Zero allocations**: Lion generates zero garbage[*](#benchmarks).

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
## Table of contents

  - [Install/Update](#installupdate)
  - [Hello World](#hello-world)
  - [Getting started with modules and resources](#getting-started-with-modules-and-resources)
  - [Handlers](#handlers)
    - [Using Handlers](#using-handlers)
    - [Using HandlerFuncs](#using-handlerfuncs)
    - [Using native http.Handler](#using-native-httphandler)
      - [Using native http.Handler with *lion.Wrap()*](#using-native-httphandler-with-lionwrap)
      - [Using native http.Handler with *lion.WrapFunc()*](#using-native-httphandler-with-lionwrapfunc)
  - [Middlewares](#middlewares)
    - [Using Named Middlewares](#using-named-middlewares)
    - [Using Negroni Middlewares](#using-negroni-middlewares)
  - [Match Hosts](#match-hosts)
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
  - [License](#license)
  - [Todo](#todo)
  - [Credits](#credits)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Install/Update

```shell
$ go get -u github.com/celrenheit/lion
```


## Hello World

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/celrenheit/lion"
	"golang.org/x/net/context"
)

func Home(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home")
}

func Hello(c context.Context, w http.ResponseWriter, r *http.Request) {
	name := lion.Param(c, "name")
	fmt.Fprintf(w, "Hello "+name)
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
```

Try it yourself. Run:
```shell
$ go run examples/modular-hello/modular-hello.go
```

Open your web browser to [http://localhost:3000/api/products](http://localhost:3000/api/products) or [http://localhost:3000/api/products/123](http://localhost:3000/api/products/123). You should see "_Fetching all products_" or "_Getting product: 123_".

## Handlers

Handlers should implement the Handler interface:

```go
type Handler interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}
```

### Using Handlers

```go
l.Get("/get", get)
l.Post("/post", post)
l.Put("/put", put)
l.Delete("/delete", delete)
```

### Using HandlerFuncs

HandlerFuncs shoud have this function signature:

```go
func handlerFunc(c context.Context, w http.ResponseWriter, r *http.Request)  {
  fmt.Fprintf(w, "Hi!")
}

l.GetFunc("/get", handlerFunc)
l.PostFunc("/post", handlerFunc)
l.PutFunc("/put", handlerFunc)
l.DeleteFunc("/delete", handlerFunc)
```


### Using native http.Handler

```go
type nativehandler struct {}

func (_ nativehandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

l.GetH("/somepath", nativehandler{})
l.PostH("/somepath", nativehandler{})
l.PutH("/somepath", nativehandler{})
l.DeleteH("/somepath", nativehandler{})
```

#### Using native http.Handler with *lion.Wrap()*

*Note*: using native http handler you cannot access url params.

```go

func main() {
	l := lion.New()
	l.Get("/somepath", lion.Wrap(nativehandler{}))
}
```

#### Using native http.Handler with *lion.WrapFunc()*


```go
func getHandlerFunc(w http.ResponseWriter, r *http.Request) {

}

func main() {
	l := lion.New()
	l.Get("/somepath", lion.WrapFunc(getHandlerFunc))
}
```

## Middlewares

Middlewares should implement the Middleware interface:

```go
type Middleware interface {
	ServeNext(Handler) Handler
}
```

The ServeNext function accepts a Handler and returns a Handler.

You can also use MiddlewareFuncs. For example:

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

To reuse it later in your application, you can use the `UseNamed` method. If it cannot find the named middleware if the current Router instance it will try to find it in the parent router.
If a named middleware is not found it will panic.

```go
api := l.Group("/api")
api.UseNamed("EnsureAuthenticated")
```

### Using Negroni Middlewares

You can use [Negroni](https://github.com/codegangsta/negroni) middlewares you can find a list of third party middlewares [here](https://github.com/codegangsta/negroni#third-party-middleware)

```go
l := lion.New()
l.UseNegroni(negroni.NewRecovery())
l.Run()
```

## Matching Hosts

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
Get(c context.Context, w http.ResponseWriter, r *http.Request)
Post(c context.Context, w http.ResponseWriter, r *http.Request)
Put(c context.Context, w http.ResponseWriter, r *http.Request)
Delete(c context.Context, w http.ResponseWriter, r *http.Request)
```

**_Example_**:

```go
package main

type todolist struct{}

func (t todolist) Uses() lion.Middlewares {
	return lion.Middlewares{lion.NewLogger()}
}

func (t todolist) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
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
func (t api) Get(c context.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This also a resource accessible at http://localhost:3000/api")
}

// Optional: Defining custom routes
func (t api) CustomRoute(c context.Context, w http.ResponseWriter, r *http.Request) {
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

# Custom Middlewares

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


### Custom Logger example

```go
type logger struct{}

func (*logger) ServeNext(next lion.Handler) lion.Handler {
	return lion.HandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
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

Without path cleaning

```
BenchmarkLion_Param       	10000000	       164 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_Param5      	 5000000	       372 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_Param20     	 1000000	      1080 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParamWrite  	10000000	       180 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GithubStatic	10000000	       160 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GithubParam 	 5000000	       359 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GithubAll   	   30000	     62888 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlusStatic 	20000000	       104 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlusParam  	10000000	       182 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlus2Params	 5000000	       286 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlusAll    	  500000	      3227 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParseStatic 	10000000	       123 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParseParam  	10000000	       145 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_Parse2Params	10000000	       212 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParseAll    	  300000	      5242 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_StaticAll   	   50000	     37998 ns/op	       0 B/op	       0 allocs/op
```

With path cleaning

```
BenchmarkLion_Param       	10000000	       227 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_Param5      	 3000000	       427 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_Param20     	 1000000	      1321 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParamWrite  	 5000000	       256 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GithubStatic	10000000	       214 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GithubParam 	 3000000	       445 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GithubAll   	   20000	     88664 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlusStatic 	10000000	       122 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlusParam  	 5000000	       381 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlus2Params	 5000000	       409 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_GPlusAll    	  500000	      3952 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParseStatic 	10000000	       146 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParseParam  	10000000	       187 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_Parse2Params	 5000000	       314 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_ParseAll    	  200000	      7857 ns/op	       0 B/op	       0 allocs/op
BenchmarkLion_StaticAll   	   30000	     56170 ns/op	      96 B/op	       8 allocs/op
```

A more in depth benchmark with a comparison with other frameworks is coming soon.

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

* @codegangsta for https://github.com/codegangsta/negroni
	* Static and Recovery middlewares are taken from Negroni
* @zenazn for https://github.com/zenazn/goji/
	* RealIP middleware is taken from goji
* @pkieltyka for https://github.com/pressly/chi and @armon for https://github.com/armon/go-radix
  * Radix tree matcher implementation is inspired by both of these packages
* https://github.com/gin-gonic/gin for ResponseWriter and Logger inspiration

package lion

import (
	"net/http"
	"strings"

	"github.com/celrenheit/lion/internal/matcher"
)

// RegisterMatcher registers and matches routes to Handlers
type registerMatcher interface {
	Register(method, pattern string, handler http.Handler) *route
	Match(*ctx, *http.Request) (*ctx, http.Handler)
	Path(pattern string, params map[string]string) (string, error)
}

////////////////////////////////////////////////////////////////////////////
///												RADIX 																				 ///
////////////////////////////////////////////////////////////////////////////

var _ registerMatcher = (*pathMatcher)(nil)

type pathMatcher struct {
	matcher matcher.Matcher
	tags    matcher.Tags
}

func newPathMatcher() *pathMatcher {
	cfg := &matcher.Config{
		ParamChar:    ':',
		WildcardChar: '*',
		Separators:   "/.",
		New: func() matcher.GetSetter {
			return &route{}
		},
	}

	r := &pathMatcher{
		matcher: matcher.Custom(cfg),
		tags:    matcher.Tags{""},
	}
	return r
}

func (d *pathMatcher) Register(method, pattern string, handler http.Handler) *route {
	d.prevalidation(method, pattern)

	rt := d.matcher.Set(pattern, handler, matcher.Tags{method})
	return rt.(*route)
}

func (d *pathMatcher) Match(c *ctx, r *http.Request) (*ctx, http.Handler) {
	p := cleanPath(r.URL.Path)

	d.tags[0] = r.Method

	h := d.matcher.GetWithContext(c, p, d.tags)

	if h == nil {
		if r.Method == OPTIONS {
			hh := d.automaticOptionsHandler(c, r.URL.Path)
			return c, hh
		}
		return c, nil
	}

	return c, h.(http.Handler)
}

func (d *pathMatcher) prevalidation(method, pattern string) {
	if len(pattern) == 0 || pattern[0] != '/' {
		panicl("path must begin with '/' in path '" + pattern + "'")
	}

	// Is http method allowed
	if !isInStringSlice(allowedHTTPMethods[:], method) {
		panicl("lion: invalid http method => %s\n\tShould be one of %v", method, allowedHTTPMethods)
	}
}

func (d *pathMatcher) automaticOptionsHandler(c *ctx, path string) http.Handler {
	allowed := make([]string, 0, len(allowedHTTPMethods))
	for _, method := range allowedHTTPMethods {
		if method == OPTIONS {
			continue
		}

		d.tags[0] = method
		h := d.matcher.GetWithContext(c, path, d.tags)
		if h != nil {
			allowed = append(allowed, method)
		}
	}

	if len(allowed) == 0 { // There is no method allowed
		return nil
	}

	allowed = append(allowed, OPTIONS)

	joined := strings.Join(allowed, ",")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept", joined)
		w.WriteHeader(http.StatusOK)
	})
}

func (d *pathMatcher) Path(pattern string, params map[string]string) (string, error) {
	return d.matcher.Eval(pattern, params)
}

func isInStringSlice(slice []string, expected string) bool {
	for _, val := range slice {
		if val == expected {
			return true
		}
	}
	return false
}

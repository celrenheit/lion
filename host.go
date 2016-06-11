package lion

import (
	"net/http"

	"github.com/celrenheit/pmatch"
)

const defaultAnyHostKey = "liondefaulthostname"

type hostMatcher struct {
	matcher pmatch.Matcher
}

func newHostMatcher() *hostMatcher {
	cfg := &pmatch.Config{
		ParamChar:        ':',
		WildcardChar:     '*',
		Separators:       "/.",
		GetSetterCreator: &hscreator{},
	}
	return &hostMatcher{
		matcher: pmatch.Custom(cfg),
	}
}

type registererRMGrabber struct {
	rm RegisterMatcher
}

func (hm *hostMatcher) Register(pattern string) RegisterMatcher {
	host := pattern
	if host == "" {
		host = "*" + defaultAnyHostKey
	}

	rg := &registererRMGrabber{}
	reversedHost := reverseHost(host)
	hm.matcher.Set(reversedHost, rg, nil)
	return rg.rm
}

func (hm *hostMatcher) Match(c *Context, req *http.Request) Handler {
	reversedHost := reverseHost(req.Host)
	value := hm.matcher.GetWithContext(c, reversedHost, nil)
	// Delete wildcard param
	// TODO: Skip this step for performance reasons
	// (Maybe by adding a blacklisted or skiplisted params on host matcher)
	if _, ok := c.ParamOk(defaultAnyHostKey); ok {
		c.Remove(defaultAnyHostKey)
	}

	if rm, ok := value.(RegisterMatcher); ok {
		_, h := rm.Match(c, req)
		return h
	}
	return nil
}

type hostStore struct {
	rm RegisterMatcher
}

func (hs *hostStore) Set(value interface{}, tags pmatch.Tags) {
	if rg, ok := value.(*registererRMGrabber); ok {
		rg.rm = hs.rm
	}
}

func (hs *hostStore) Get(tags pmatch.Tags) interface{} {
	return hs.rm
}

type hscreator struct{}

func (c *hscreator) New() pmatch.GetSetter {
	return &hostStore{
		rm: newRadixMatcher(),
	}
}

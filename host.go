package lion

import (
	"net/http"
	"strings"
	"sync"

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
		Separators:       ".",
		GetSetterCreator: &hscreator{},
		ParamTransformer: newHostParamTransformer(),
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

type hostParamTransformer struct {
	splittedStringPool sync.Pool
}

func newHostParamTransformer() *hostParamTransformer {
	return &hostParamTransformer{
		splittedStringPool: sync.Pool{
			New: func() interface{} {
				return &splittedStringItem{}
			},
		},
	}
}

func (hpt *hostParamTransformer) Transform(input string) string {
	reversedItem := hpt.split(input, ".")
	reversed := reversedItem.slice
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	output := strings.Join(reversed, ".")
	hpt.splittedStringPool.Put(reversedItem)
	return output
}

// Taken from Go's standard library
// https://github.com/golang/go/blob/master/src/strings/strings.go#L237-L261
func (hpt *hostParamTransformer) split(s, sep string) *splittedStringItem {
	si := hpt.splittedStringPool.Get().(*splittedStringItem)
	n := strings.Count(s, sep) + 1
	c := sep[0]
	start := 0

	var a []string
	if cap(si.slice) == n {
		a = si.slice
	} else {
		a = make([]string, n)
	}

	na := 0
	for i := 0; i+len(sep) <= len(s) && na+1 < n; i++ {
		if s[i] == c && (len(sep) == 1 || s[i:i+len(sep)] == sep) {
			a[na] = s[start:i]
			na++
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	a[na] = s[start:]
	si.slice = a[0 : na+1]
	return si
}

type splittedStringItem struct {
	slice []string
}

var hostReverser = newHostParamTransformer()

func reverseHost(input string) string {
	return hostReverser.Transform(input)
}

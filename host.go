package lion

import (
	"net/http"
	"strings"

	"github.com/celrenheit/lion/internal/matcher"
)

const (
	defaultAnyHostKey     = "lionDefaultAnyHostKey"
	defaultAnyHostPattern = "*" + defaultAnyHostKey
)

type hostMatcher struct {
	matcher   matcher.Matcher
	defaultRM registerMatcher
	multihost bool
}

func newHostMatcher() *hostMatcher {
	cfg := &matcher.Config{
		ParamChar:    '$',
		WildcardChar: '*',
		Separators:   ".:",
		New: func() matcher.Store {
			return &hostStore{
				rm: newPathMatcher(),
			}
		},
		ParamTransformer: newHostParamTransformer(),
	}
	return &hostMatcher{
		matcher:   matcher.Custom(cfg),
		defaultRM: newPathMatcher(),
	}
}

func (hm *hostMatcher) Register(pattern string) registerMatcher {
	host := pattern

	// Switch to multihost
	if !hm.multihost && host != "" {
		hm.multihost = true
		hm.matcher.Set(reverseHost(defaultAnyHostPattern), hm.defaultRM, nil)
	}

	if hm.multihost {

		if host == "" {
			host = defaultAnyHostPattern
		}
		hs := &hostStore{}
		reversedHost := reverseHost(host)
		hs = hm.matcher.Set(reversedHost, hs, nil).(*hostStore)
		return hs.rm
	}

	return hm.defaultRM
}

func (hm *hostMatcher) Match(c *ctx, req *http.Request) http.Handler {
	if hm.multihost {
		reversedHost := reverseHost(req.Host)
		value, _ := hm.matcher.GetWithContext(c, reversedHost, nil)
		// Delete wildcard param
		// TODO: Skip this step for performance reasons
		// (Maybe by adding a blacklisted or skiplisted params on host matcher)
		if _, ok := c.ParamOk(defaultAnyHostKey); ok {
			c.Remove(defaultAnyHostKey)
		}

		if rm, ok := value.(registerMatcher); ok {
			_, h := rm.Match(c, req)
			return h
		}
	} else {
		_, h := hm.defaultRM.Match(c, req)
		return h
	}
	return nil
}

type hostStore struct {
	rm registerMatcher
}

func (hs *hostStore) Set(value interface{}, tags matcher.Tags) {
	// Overwrite RegisterMatcher
	if rm, ok := value.(registerMatcher); ok {
		hs.rm = rm
	}
}

func (hs *hostStore) Get(tags matcher.Tags) interface{} {
	return hs.rm
}

type hostParamTransformer struct{}

func newHostParamTransformer() *hostParamTransformer {
	return &hostParamTransformer{}
}

func (hpt *hostParamTransformer) Transform(input string) string {
	// Split host based on '.' character
	reversed := hpt.split(input, ".")

	// Split and reverse each host parts if it has a port character ':'
	portPart := reversed[len(reversed)-1]
	if strings.Contains(portPart, ":") {
		splitted := strings.Split(portPart, ":")
		splitted[0], splitted[1] = splitted[1], splitted[0]
		reversed[len(reversed)-1] = strings.Join(splitted, ":")
	}

	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	output := strings.Join(reversed, ".")
	return output
}

// Taken from Go's standard library
// https://github.com/golang/go/blob/master/src/strings/strings.go#L237-L261
func (hpt *hostParamTransformer) split(s, sep string) []string {
	slice := []string{}
	n := strings.Count(s, sep) + 1
	c := sep[0]
	start := 0

	var a []string
	if cap(slice) == n {
		a = slice
	} else {
		a = make([]string, n)
	}

	na := 0
	for i := 0; i+len(sep) <= len(s) && na+1 < n; i++ {
		if s[i] == c && (len(sep) == 1 || s[i:i+len(sep)] == sep) && (i == 0 || s[i-1] != '\\') {
			a[na] = s[start:i]
			na++
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	a[na] = s[start:]
	slice = a[0 : na+1]
	return slice
}

var hostReverser = newHostParamTransformer()

func reverseHost(input string) string {
	return hostReverser.Transform(input)
}

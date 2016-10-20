package lion

import (
	"net/http"
	"testing"

	"context"
)

// mss is an alias for map[string]string.
type mss map[string]string

func (p *ctx) toMap() mss {
	m := mss{}
	for i := range p.keys {
		m[p.keys[i]] = p.values[i]
	}
	return m
}

func TestContextAddParam(t *testing.T) {
	c := newContext()
	c.parent = context.WithValue(context.TODO(), "parentKey", "parentVal")
	c.AddParam("key", "val")
	if len(c.keys) != 1 {
		t.Errorf("Length of keys should be 1 but got %s", red(len(c.keys)))
	}

	val := c.Value("key")
	if val == nil {
		t.Errorf("Context Value() should not be nil")
	}
	str, ok := val.(string)
	if !ok {
		t.Error("Context value is not a string")
	}

	if str != "val" {
		t.Error("Wrong value for Context.Value()")
	}

	parentVal := c.Value("parentKey")
	if parentVal == nil {
		t.Errorf("Context parent Value() should not be nil")
	}
	str, ok = parentVal.(string)
	if !ok {
		t.Error("Context parent value is not a string")
	}

	if str != "parentVal" {
		t.Error("Wrong value for parent Context.Value()")
	}
}

func TestContextC(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	c := C(req)
	ctx := c.(*ctx)
	if ctx.parent != context.Background() {
		t.Error("Context C: Parent should be context.TODO()")
	}
}

func TestGetParamInNestedContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContextWithParent(context.Background())
	c.AddParam("id", "myid")
	req = req.WithContext(c)

	base := context.WithValue(context.Background(), ctxKey, c)
	nc := context.WithValue(base, "db", "mydb")
	nc = context.WithValue(nc, "t", "t")
	req = req.WithContext(nc)

	testc := req.Context()
	pc := testc.Value(ctxKey)
	if pc == nil {
		t.Errorf("Should not be nil")
	}

	if _, ok := pc.(*ctx); !ok {
		t.Errorf("Should be a *ctx")
	}

	id := Param(req, "id")
	if id != "myid" {
		t.Errorf("id should be equal to 'myid' but got '%s'", id)
	}

	nonexistant := Param(req, "nonexistant")
	if nonexistant != "" {
		t.Errorf("Should be empty but got '%s'", nonexistant)
	}

	val := testc.Value("db")
	v, ok := val.(string)
	if !ok {
		t.Errorf("Should be a string")
	}

	if v != "mydb" {
		t.Errorf("Value should be equal to 'mydb' but got '%s'", v)
	}
}

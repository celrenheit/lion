package lion

import (
	"testing"

	"golang.org/x/net/context"
)

func TestContextAddParam(t *testing.T) {
	c := NewContext()
	c.parent = context.WithValue(context.TODO(), "parentKey", "parentVal")
	c.addParam("key", "val")
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
	c := C(context.TODO())
	if c.parent != context.TODO() {
		t.Error("Context C: Parent should be context.TODO()")
	}
}

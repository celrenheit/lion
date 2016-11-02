package lion

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"reflect"
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

	c := newContext()
	c.AddParam("id", "myid")
	c.parent = req.Context()

	nc := context.WithValue(c, "db", "mydb")
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

func TestContextClone(t *testing.T) {
	old := newContextWithParent(context.Background())
	old.AddParam("test", "val")
	new := old.Clone().(*ctx)

	if !reflect.DeepEqual(old, new) {
		t.Errorf("Should be equal")
	}
}

func TestContextRender(t *testing.T) {
	tests := map[string][]struct {
		input    interface{}
		expected string
	}{
		"string": {
			{
				input:    "Hello",
				expected: "Hello",
			},
		},
		"json": {
			{
				input: map[string]interface{}{
					"test": map[string]string{
						"test2": "val",
					},
				},
				expected: `{"test":{"test2":"val"}}` + "\n",
			},
		},
		"xml": {
			{
				input: &struct {
					XMLName   xml.Name `xml:"person"`
					Id        int      `xml:"id,attr"`
					FirstName string   `xml:"name>first"`
					LastName  string   `xml:"name>last"`
					Age       int      `xml:"age"`
					Height    float32  `xml:"height,omitempty"`
					Married   bool
					Comment   string `xml:",comment"`
				}{Id: 13, FirstName: "John", LastName: "Doe", Age: 42},
				expected: `<person id="13"><name><first>John</first><last>Doe</last></name><age>42</age><Married>false</Married></person>`,
			},
		},
	}

	for dtype, subtests := range tests {
		t.Run(dtype, func(t *testing.T) {
			for _, test := range subtests {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/hello", nil)
				c := newContextWithResReq(context.Background(), w, r)

				switch dtype {
				case "json":
					c.JSON(test.input)
				case "xml":
					c.XML(test.input)
				case "string":
					c.String(test.input.(string))
				default:
					panicl("unsupported test %s", dtype)
				}

				got := w.Body.String()
				want := test.expected

				if got != want {
					t.Errorf("Expected '%s' but got '%s'", want, got)
				}
			}
		})
	}
}

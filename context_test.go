package lion

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
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
		input       interface{}
		expected    string
		shouldError bool
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
				expected: `{"test":{"test2":"val"}}`,
			},
			{
				input:       math.Inf(1),
				expected:    "",
				shouldError: true,
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
				t.Run(fmt.Sprintf("%v", test.input), func(t *testing.T) {
					c, w := newTestCtx()

					var ctype string
					var err error
					switch dtype {
					case "json":
						err = c.JSON(test.input)
						ctype = contentTypeJSON
					case "xml":
						err = c.XML(test.input)
						ctype = contentTypeXML
					case "string":
						err = c.String(test.input.(string))
						ctype = contentTypeTextPlain
					default:
						panicl("unsupported test %s", dtype)
					}

					if test.shouldError {
						if err == nil {
							t.Error("Should error")
						}
						return
					}

					got := w.Body.String()
					want := test.expected

					if got != want {
						t.Errorf("Expected '%s' but got '%s'", want, got)
					}

					gotType := w.Header().Get("Content-Type")
					wantType := ctype

					if gotType != wantType {
						t.Errorf("Expected '%s' but got '%s'", wantType, gotType)
					}
				})
			}
		})
	}
}

func TestContextFile(t *testing.T) {
	want := "Lion"

	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	f.WriteString(want)
	f.Close()

	defer os.Remove(f.Name())

	c, w := newTestCtx()

	c.File(f.Name())

	got := w.Body.String()
	if got != want {
		t.Errorf("Expected '%s' but got '%s'", want, got)
	}

	want = "text/plain; charset=utf-8"
	got = w.Header().Get("Content-Type")
	if got != want {
		t.Errorf("Expected '%s' but got '%s'", want, got)
	}

	c, w = newTestCtx()

	f2, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	f2.WriteString(`<!DOCTYPE html><html><body><div class="lion">arghhh</div></body></html>`)
	f2.Close()

	defer os.Remove(f2.Name())
	c.File(f2.Name())

	want = "text/html; charset=utf-8"
	got = w.Header().Get("Content-Type")
	if got != want {
		t.Errorf("Expected '%s' but got '%s'", want, got)
	}
}

func TestWithStatus(t *testing.T) {
	c, w := newTestCtx()

	c.WithStatus(200)
	if w.Code != 200 {
		t.Errorf("Context: status code should be equal")
	}

	c.WithStatus(401)
	if w.Code != 200 {
		t.Errorf("Context: status code should not have changed")
	}
}

func TestWithHeader(t *testing.T) {
	c, w := newTestCtx()

	c.WithHeader("test", "val")

	if w.Header().Get("test") != "val" {
		t.Errorf("Context: header should be equal")
	}

}

func TestDoNotWriteStatusIfErr(t *testing.T) {
	c, w := newTestCtx()
	err := c.WithStatus(500).JSON(math.Inf(-1))
	if w.Code != 200 {
		t.Errorf("Default code should be 200 OK because it should not be written")
	}

	if err == nil {
		t.Errorf("Invalid JSON: Should have an err != nil")
	}

	err = c.WithStatus(401).JSON(map[string]interface{}{
		"test": map[string]string{
			"test2": "val",
		},
	})

	if w.Code != 401 {
		t.Errorf("Return code should be 401 but got %s", red(w.Code))
	}
}

func TestRedirect(t *testing.T) {
	c, w := newTestCtx()
	c.WithStatus(http.StatusTemporaryRedirect).
		Redirect("/goodbye")

	want := "/goodbye"
	got := w.Header().Get("Location")
	if got != want {
		t.Errorf("Expected '%s' but got '%s'", want, got)
	}

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Should have status %d but got %d", http.StatusTemporaryRedirect, w.Code)
	}
}

func newTestCtx() (c *ctx, w *httptest.ResponseRecorder) {
	w = httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/hello", nil)
	c = newContextWithResReq(context.Background(), w, r)

	return
}

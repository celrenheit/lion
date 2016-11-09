package lion

import "testing"

func TestRouteGeneratePath(t *testing.T) {
	l := New()
	register := []struct {
		pattern, name string
	}{
		{"/a/:name", "a_name"},
		{"/a/:name/:n([0-9]+)", "a_name_n"},
		{"/a/b/:dest/*path", "a_b_dest_path"},
		{"/e/:file.:ext", "e_file_ext"},
	}
	for _, r := range register {
		l.Get(r.pattern, fakeHandler()).WithName(r.name)
	}

	tests := []struct {
		route_name   string
		params       map[string]string
		expectedPath string
		expectedErr  bool
	}{
		{route_name: "a_name", params: mss{"name": "batman"}, expectedPath: "/a/batman"},
		{route_name: "a_name_n", params: mss{"name": "batman", "n": "123"}, expectedPath: "/a/batman/123"},
		{route_name: "a_name_n", params: mss{"name": "batman", "n": "1d23"}, expectedErr: true},
		{route_name: "a_b_dest_path", params: mss{"dest": "batman", "path": "subfolder/test/hello.jpeg"}, expectedPath: "/a/b/batman/subfolder/test/hello.jpeg"},
		{route_name: "e_file_ext", params: mss{"file": "test", "ext": "mp4"}, expectedPath: "/e/test.mp4"},
	}

	for _, test := range tests {
		path, err := l.Route(test.route_name).Path(test.params)
		if test.expectedErr && err == nil {
			t.Errorf("Should have errored")
		}

		if !test.expectedErr && err != nil {
			t.Error(err)
		}

		if path != test.expectedPath {
			t.Errorf("Incorrect path: got '%s' want '%s'", path, test.expectedPath)
		}
	}
}

func TestGetRoutesSubrouter(t *testing.T) {
	l := New()
	l.Get("/hello", fakeHandler())
	api := l.Group("/api")
	api.Get("/users", fakeHandler())
	api.Get("/posts", fakeHandler())
	api.Get("/sessions", fakeHandler())

	got := len(l.Routes())
	if got != 4 {
		t.Errorf("Number of routes should be 4 but got %d: %v", got, l.Routes())
	}

	got = len(api.Routes())
	if got != 3 {
		t.Errorf("Number of routes should be 3 but got %d: %v", got, api.Routes())
	}

	lv2 := New()
	lv2.Get("/users2", fakeHandler())
	lv2.Get("/posts2", fakeHandler())
	lv2.Get("/sessions2", fakeHandler())

	l.Mount("/v2", lv2)

	got = len(l.Routes())
	if got != 7 {
		t.Errorf("Number of routes should be 7 but got %d: %v", got, l.Routes())
	}

	lv2.Get("/categories", fakeHandler())
	got = len(l.Routes())
	if got != 8 {
		t.Errorf("Number of routes should be 8 but got %d: %v", got, l.Routes())
	}
}

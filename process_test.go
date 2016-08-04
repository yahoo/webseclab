package webseclab

import (
	"net/http"
	"testing"
)

// mock for http.ResponseWriter is in common_test.go

// create a type for brevity and to avoid repetition in testing two similar functions
type dohandler func(w http.ResponseWriter, fpath string, input *InData) (err error)

// read the templates
func init() {
	err := parseTemplates()
	if err != nil {
		panic(err)
	}
}

// Test that doHtmlTemplate sets correct content-type
func TestTemplateContentType(t *testing.T) {
	t.Parallel()
	funcs := [...]dohandler{doHTMLTemplate, doTextTemplate}
	funcnames := [...]string{`doHtmlTemplate`, `doTextTemplate`}
	for i, f := range funcs {
		w := newMockResponseWriter()
		input := &InData{}
		err := f(w, "xss/reflect/basic", input)
		if err != nil {
			t.Errorf("Error in %s: %s\n", funcnames[i], err)
			return
		}
		if ctype := w.Header().Get("Content-Type"); ctype != "text/html; charset=utf-8" {
			t.Errorf("Unexpected ctype in %s %s - want text/html; charset=utf-8\n", funcnames[i], ctype)
		}
	}
}

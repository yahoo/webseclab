// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"strings"
	"testing"
)

var samples = map[string]string{"plain string": `plain`,
	"quotes":             `single quote: ', double quote: "abc", together: '"'`,
	"string with spaces": `plain with spaces`,
	"double quotes":      `double-quotes: ""`,
	"single quotes":      `single quote: '`,
	"tags":               `some text <xss> more text`,
	"less-than":          `matrix A < B`,
	"greater-than":       `matrix A >= B`,
	"script":             `<script>alert(xss)</script>`,
	"parens":             `onerror=alert(12345)`,
}

func TestUnescapeUnicode(t *testing.T) {
	t.Parallel()
	s := `\u0022\u003e\u003cscript\u003ealert(123)\u3c\u2fscript\u3e`
	want := `"><script>alert(123)</script>`
	esc := UnescapeUnicode(s)
	if esc != want {
		t.Errorf("Want %s, got: %s\n", want, esc)
	}
}

func TestPercentToSlash(t *testing.T) {
	t.Parallel()
	src := `%5Cx5c%5Cx5c%5Cx27%5Cx5c%5Cx5c%5Cx22`
	esc := "/x5c/x5c/x27/x5c/x5c/x22"
	res := percentToSlash(src)
	if res != esc {
		t.Errorf("Want %s, got %s\n", esc, res)
	}
}

func TestUnescapeToHex(t *testing.T) {
	t.Parallel()
	src := "http://example.com\x5c\x5c\x27\x5c\x5c\x22"
	esc := "http://example.com/x5c/x5c/x27/x5c/x5c/x22"
	res := unescapeToHex(src)
	if res != esc {
		t.Errorf("Want %s, got %s\n", esc, res)
	}
}

// check that a Transformer with no params does not change the input
func TestNewTransformerNop(t *testing.T) {
	t.Parallel()
	tr := NewTransformer()
	for k, v := range samples {
		if v != tr.Transform(v) {
			t.Errorf("Error in transform of %s: expecting nop (%s), got the change: %s\n", k, v, tr.Transform(v))
		}
	}
}

func TestTransformerQuotesOff(t *testing.T) {
	t.Parallel()
	tname := "double and single quotes 1"
	tr := NewTransformer(QuotesOff)
	res := tr.Transform(samples[tname])
	if strings.ContainsRune(res, 0x22) || strings.ContainsRune(res, 0x27) {
		t.Errorf("Error in transform of %s: quotes are not replaced in the result [[%s]]: original %s\n", tname, tr, samples[tname])
	}
}

func TestTransformerTagsOff(t *testing.T) {
	t.Parallel()
	tr := NewTransformer(TagsOff)
	for _, k := range []string{"tags", "less-than", "greater-than", "script"} {
		tname := k
		v := samples[k]
		res := tr.Transform(v)
		if strings.ContainsRune(res, 0x3c) || strings.ContainsRune(res, 0x3e) {
			t.Errorf("Error in transform of %s: no wanted replacedment in the result [[%s]] (original: %s)\n", tname, res, v)
		}
	}
}

func TestTransformerQuotesCook(t *testing.T) {
	t.Parallel()
	tr := NewTransformer(QuotesCook)
	tname := "quotes"
	res := tr.Transform(samples[tname])
	if res != `single quote: &#39;, double quote: &quot;abc&quot;, together: &#39;&quot;&#39;` {
		t.Errorf("Error in transform of %s: quotes are not properly cooked [[%s]]: original %s\n", tname, res, samples[tname])
	}
}

func TestTransformerTagsCook(t *testing.T) {
	t.Parallel()
	tr := NewTransformer(TagsCook)
	tname := "tags"
	res := tr.Transform(samples[tname])
	if res != `some text &lt;xss&gt; more text` {
		t.Errorf("Error in transform of %s: tags are not properly cooked [[%s]]: original %s\n", tname, res, samples[tname])
	}
}

func TestBackslashEscape(t *testing.T) {
	t.Parallel()
	tr := NewTransformer(BackslashEscape)
	s := `str with backslash\r\n \" double: \\ xyz`
	want := `str with backslash\\r\\n \\" double: \\\\ xyz`
	if res := tr.Transform(s); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestQuotesBackslashQuoteFullEscape(t *testing.T) {
	t.Parallel()
	tr := NewTransformer(DoubleQuotesBackslashEscape, BackslashEscape)
	s := `str with "quotes" and \ backslash`
	want := `str with \\"quotes\\" and \\ backslash`
	if res := tr.Transform(s); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestScriptOff(t *testing.T) {
	t.Parallel()
	s := `str<script>alert(123)</script>`
	want := `stralert(123)`
	if res := StripScript(s); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestScriptOffWithAttr(t *testing.T) {
	t.Parallel()
	s := `str<script id="foo">alert(123)</script>`
	want := `stralert(123)`
	if res := StripScript(s); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestNestedTransformer(t *testing.T) {
	t.Parallel()
	s := `str<script id="foo">alert('123')<xss></script>`
	want := `stralert(123)<xss>`
	tr1 := NewTransformer(QuotesOff)
	tr2 := NewTransformerRegexp(`</?script[^>]*>`)
	rt, ok := tr2.(RegexpTransformer)
	if ok != true {
		t.Errorf("In TestNestedTransformer, cannot convert Transformer to RegexpTransformer")
		return
	}
	rt.nested = &tr1
	if res := rt.Transform(s); res != want {
		t.Errorf("Error in the nested transform test: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestParensOff(t *testing.T) {
	t.Parallel()
	s := `onerror=alert(12345)`
	want := `onerror=alert12345`
	tr := NewTransformer(ParensOff)
	if res := tr.Transform(s); res != want {
		t.Errorf("parens stripping: want %s, got %s (orig: %s)\n", want, res, s)
	}
}

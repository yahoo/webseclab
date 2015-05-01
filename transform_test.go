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

func TestNoOp(t *testing.T) {
	for label, s := range samples {
		if res := Transform(s, NoOp); res != s {
			t.Errorf("Want %s (identity), got %s in %s\n", s, res, label)
		}
	}
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

func TestTransformerQuotesOff(t *testing.T) {
	t.Parallel()
	tname := "double and single quotes 1"
	res := Transform(samples[tname], QuotesOff)
	if strings.ContainsRune(res, 0x22) || strings.ContainsRune(res, 0x27) {
		t.Errorf("Error in transform of %s: quotes are not replaced in the result [[%s]]: original %s\n", tname, res, samples[tname])
	}
}

func TestRemoveTags(t *testing.T) {
	t.Parallel()
	s := `hi<xss> nasty tags<script src="foo"> abound<b>around</b>!`
	w := `hi  nasty tags  abound around !`
	if res := Transform(s, TagsOff); res != w {
		t.Errorf("tags off except textarea close:\nwant %s\ngot  %s\n(orig: %s)\n", w, res, s)
	}
}

func TestTransformerTagsOff(t *testing.T) {
	t.Parallel()
	for _, k := range []string{"tags", "less-than", "greater-than", "script"} {
		tname := k
		v := samples[k]
		res := Transform(v, TagCharsOff)
		if strings.ContainsRune(res, 0x3c) || strings.ContainsRune(res, 0x3e) {
			t.Errorf("Error in transform of %s: no wanted replacement in the result [[%s]] (original: %s)\n", tname, res, v)
		}
	}
}

func TestTransformerQuotesCook(t *testing.T) {
	t.Parallel()
	tname := "quotes"
	res := Transform(samples[tname], QuotesCook)
	if res != `single quote: &#39;, double quote: &quot;abc&quot;, together: &#39;&quot;&#39;` {
		t.Errorf("Error in transform of %s: quotes are not properly cooked [[%s]]: original %s\n", tname, res, samples[tname])
	}
}

func TestTransformerTagsCook(t *testing.T) {
	t.Parallel()
	tname := "tags"
	res := Transform(samples[tname], TagsCook)
	if res != `some text &lt;xss&gt; more text` {
		t.Errorf("Error in transform of %s: tags are not properly cooked [[%s]]: original %s\n", tname, res, samples[tname])
	}
}

func TestBackslashEscape(t *testing.T) {
	t.Parallel()
	s := `str with backslash\r\n \" double: \\ xyz`
	want := `str with backslash\\r\\n \\" double: \\\\ xyz`
	if res := Transform(s, BackslashEscape); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestQuotesBackslashQuoteFullEscape(t *testing.T) {
	t.Parallel()
	s := `str with "quotes" and \ backslash`
	want := `str with \\"quotes\\" and \\ backslash`
	if res := Transform(s, DoubleQuotesBackslashEscape, BackslashEscape); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestBackslashDoublequotesEscape(t *testing.T) {
	t.Parallel()
	s := `str with "quotes" and \ backslash and both \"`
	want := `str with \"quotes\" and \\ backslash and both \\\"`
	if res := Transform(s, BackslashEscapeDoubleQuotesAndBackslash); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}
func TestScriptOff(t *testing.T) {
	t.Parallel()
	s := `str<script>alert(123)</script>`
	want := `stralert(123)`
	if res := Transform(s, ScriptOff); res != want {
		t.Errorf("Error in transform: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestScriptOffWithAttr(t *testing.T) {
	t.Parallel()
	s := `str<script id="foo">alert(123)</script>`
	want := `stralert(123)`
	if res := Transform(s, ScriptOff); res != want {
		t.Errorf("Error in TestScriptOffWithAttr: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestTextareaClose(t *testing.T) {
	t.Parallel()
	inp := [...]string{`str<textarea>foo</textarea>`, `str<textarea>foo</TEXTAREA>`, `str<textarea>foo</TeXtArEa>`}
	want := `str<textarea>foo`
	for _, s := range inp {
		if res := Transform(s, TextareaCloseOff); res != want {
			t.Errorf("Error in StripTextareaClose: want %s got %s; original: %s\n", want, res, s)
		}
	}
}

func TestQuotesOffScriptOff(t *testing.T) {
	t.Parallel()
	s := `str<script id="foo">alert('123')<xss></script>`
	want := `stralert(123)<xss>`
	if res := Transform(s, QuotesOff, ScriptOff); res != want {
		t.Errorf("Error in the TestQuotesOffScriptOff test: want %s got %s; original: %s\n", want, res, s)
	}
}

func TestParensOff(t *testing.T) {
	t.Parallel()
	s := `onerror=alert(12345)`
	want := `onerror=alert12345`
	if res := Transform(s, ParensOff); res != want {
		t.Errorf("parens stripping: want %s, got %s (orig: %s)\n", want, res, s)
	}
}

func TestRemoveTagsUntilTextareaClose(t *testing.T) {
	t.Parallel()
	s := `name="in" rows="5" cols="60">"><script>alert("xss");</script>xsrcin</textarea>foo<xss>bar`
	w := `name="in" rows="5" cols="60">">alert("xss");xsrcin</textarea>foo<xss>bar`
	if res := RemoveTagsUntilTextareaClose(s); res != w {
		t.Errorf("tags off except textarea close:\nwant %s\ngot  %s\n(orig: %s)\n", w, res, s)
	}
}

func TestTagsOffExceptTextareaClose(t *testing.T) {
	t.Parallel()
	s := `name="in" rows="5" cols="60">"><script>alert("xss");</script>xsrcin</textarea>foo<xss>bar`
	w := `name="in" rows="5" cols="60">"> alert("xss"); xsrcin</textarea>foo bar`
	if res := Transform(s, TagsOffExceptTextareaClose); res != w {
		t.Errorf("tags off except textarea close:\nwant %s\ngot  %s\n(orig: %s)\n", w, res, s)
	}
}

func TestTextareaSafe(t *testing.T) {
	t.Parallel()
	s := `name="in" rows="5" cols="60">"><script>alert("xss");</script>xsrcin</textarea>foo<xss>danger`
	w := `name="in" rows="5" cols="60">"><script>alert("xss");</script>xsrcin</textarea>foo danger`
	if res := Transform(s, TextareaSafe); res != w {
		t.Errorf("textarea safe filter:\nwant %s\ngot  %s\n(orig: %s)\n", w, res, s)
	}
}

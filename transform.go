// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

// This file has two parts - one related to the map of
// filter fields to the replacers (for more standard replacements).
// The second part are the functions that do transformations
// beyond simple string substitution (regexp etc.)

// Transformer transforms a string by escaping, filtering or other modification.
type Transformer interface {
	Transform(s string) string
}

// StringsReplacer implements Transformer using embedded strings.Replacer.
type StringsReplacer struct {
	*strings.Replacer
}

// NewStringsReplacer creates a new StringsReplacer
// using the list of old/new strings (as in strings.NewReplacer).
func NewStringsReplacer(oldnew ...string) *StringsReplacer {
	r := strings.NewReplacer(oldnew...)
	return &StringsReplacer{r}
}

// Transform implements the Transformer interface by calling
// string replacement function.
func (r *StringsReplacer) Transform(s string) string {
	return r.Replace(s)
}

// RegexpMatchEraser implements Tranformer using the given regexp(s).
type RegexpMatchEraser struct {
	// slice to allow multiple regexps for removing strings
	re []*regexp.Regexp
}

// NewRegexpMatchEraser accepts a regexp string parameter
// returns a Transformer that removes the matching strings.
func NewRegexpMatchEraser(re ...string) *RegexpMatchEraser {
	var r []*regexp.Regexp
	for _, pattern := range re {
		compiled := regexp.MustCompile(pattern)
		r = append(r, compiled)
	}
	return &RegexpMatchEraser{r}
}

// Transform erases matching strings based on embedded regexp(s).
func (r *RegexpMatchEraser) Transform(s string) string {
	for _, p := range r.re {
		s = p.ReplaceAllLiteralString(s, "")
	}
	return s
}

// ReplaceFunction is the type alias for the Transformer interface.
type ReplaceFunction func(string) string

// Transform satisfies the Transformer interface by
// applying the functor on the string parameter.
func (f ReplaceFunction) Transform(s string) string {
	return f(s)
}

// identity function
func id(s string) string {
	return s
}

type transformerMap map[filter]Transformer

var trMap transformerMap

// initialize the main replacing functions
func init() {
	trMap = transformerMap(make(map[filter]Transformer))
	trMap[BackslashEscape] = NewStringsReplacer(`\`, `\\`)
	trMap[DoubleQuotesBackslashEscape] = NewStringsReplacer(`"`, `\"`)
	trMap[BackslashEscapeDoubleQuotesAndBackslash] = ReplaceFunction(backslashDoublequotes)
	trMap[DoubleQuotesCook] = NewStringsReplacer(`"`, `&quot;`)
	trMap[DoubleQuotesOff] = NewStringsReplacer(`"`, "")
	trMap[GreaterThanCook] = NewStringsReplacer(`>`, `&gt;`)
	trMap[GreaterThanOff] = NewStringsReplacer(`>`, "")
	trMap[LessThanCook] = NewStringsReplacer(`<`, `&lt;`)
	trMap[LessThanOff] = NewStringsReplacer(`<`, "")
	trMap[NoOp] = ReplaceFunction(id)
	trMap[ParensOff] = NewStringsReplacer(`(`, "", `)`, "")
	trMap[QuotesCook] = NewStringsReplacer(`"`, `&quot;`, `'`, `&#39;`)
	trMap[QuotesOff] = NewStringsReplacer(`"`, "", `'`, "")
	trMap[ScriptOff] = NewRegexpMatchEraser(`(?i)<script[^>]*>`, `</script>`)
	trMap[SingleQuotesCook] = NewStringsReplacer(`'`, `&#39;`)
	trMap[SingleQuotesOff] = NewStringsReplacer(`'`, "")
	trMap[SpacesCook] = NewStringsReplacer(` `, "&#20;")
	trMap[SpacesOff] = NewStringsReplacer(` `, "")
	trMap[TagsCook] = NewStringsReplacer(`<`, `&lt;`, `>`, `&gt;`)
	trMap[TagCharsOff] = NewStringsReplacer(`<`, "", `>`, "")
	trMap[TagsOff] = ReplaceFunction(RemoveTags)
	trMap[TagsOffExceptTextareaClose] = ReplaceFunction(RemoveTagsExceptTextareaClose)
	trMap[TagsOffUntilTextareaClose] = ReplaceFunction(RemoveTagsUntilTextareaClose)
	trMap[TextareaCloseOff] = ReplaceFunction(removeTextareaClose)
	trMap[TextareaSafe] = ReplaceFunction(ReplaceTextareaSafe)
}

// Transform tranforms the string based on the given filter options (one or several)
func Transform(s string, f ...filter) string {
	if len(f) == 0 {
		log.Printf("ERROR in Tranform(%s) - empty filter slice passed!\n", s)
	}
	for _, opt := range f {
		if tr, ok := trMap[opt]; !ok {
			log.Printf("ERROR in Transform(%s, %v) - option %v is not in the trMap! Skipping.\n", s, f, opt)
			continue
		} else {
			s = tr.Transform(s)
		}
	}
	return s
}

// UnescapeUnicode takes a string with Unicode escape sequences  \u22
// and converts all of them to the unescaped characters:
// \u0022 => '"', \u3e => '>'
func UnescapeUnicode(s string) string {
	if len(s) < 4 {
		return s
	}
	re := regexp.MustCompile(`(\\u(00)?[0-9a-fA-F][0-9a-fA-F])`)
	esc := re.ReplaceAllStringFunc(s, unescapeUnicodeHelper)
	return esc
}

// same as above but for a single escape sequence
func unescapeUnicodeHelper(s string) string {
	if len(s) < 4 {
		log.Printf("ERROR in unescapeUnicode - %s is too short (< 4 chars)\n", s)
		return s
	}
	i, err := strconv.ParseInt(s[len(s)-2:], 16, 8)
	if err != nil {
		log.Printf("ERROR in unescapeUnicode(%s) - unable to ParseInt: %s\n", s, err)
		return s
	}
	if i >= 128 {
		log.Printf("ERROR in unescapeUnicode(%s) - parsed value >= 128 %d\n", s, i)
		return s
	}
	return string(rune(i))
}

func percentToSlash(s string) string {
	return strings.Replace(s, `%5C`, `/`, -1)
}

// unescapeToHex converts backslash-encoded chars \x5c, \x27 and \x22
// to strings x5c, x27, and x22
func unescapeToHex(s string) (ret string) {
	ret = ""
	for i := 0; i < len(s); i++ {
		if s[i] == 0x5c || s[i] == 0x27 || s[i] == 0x22 {
			ret += "/x"
			sOut := strconv.FormatInt(int64(s[i]), 16)
			ret += sOut
		} else {
			ret += string(s[i])
		}
	}
	return
}

// RemoveTags removes the tags: foo<xss x=1>bar => foobar
func RemoveTags(src string) string {
	re := regexp.MustCompile(`(?i)(<([^>]+)>)`)
	// be paranoid and do replacement recursevly, just in case
	for {
		copy := src
		src = re.ReplaceAllString(src, " ")
		if src == copy {
			break
		}
	}
	return re.ReplaceAllString(src, " ")
}

// RemoveTagsExceptTextareaClose removes all the tags except the closing textarea one
func RemoveTagsExceptTextareaClose(src string) (out string) {
	re := regexp.MustCompile(`(?i)(<([^>]+)>)`)
	m := re.FindAllStringIndex(src, -1)
	last := 0
	// revisit once https://github.com/golang/go/issues/5690 is resolved
	for i := 0; i < len(m); i++ {
		// copy the substring from "last" until the current match
		// - otherwise, skip the tag and add a space
		out += src[last:m[i][0]]
		if strings.HasPrefix(strings.ToLower(src[m[i][0]:m[i][1]]), "</textarea>") {
			out += src[m[i][0]:m[i][1]]
		} else {
			out += " "
		}
		last = m[i][1]
		// leave for debugging
		// fmt.Printf("%d %s\n", i, src[m[i][0]:m[i][1]])
	}
	if last < len(src) {
		out += src[last:]
	}
	// leave for debugging
	// fmt.Printf("RemoveTagsUntilTextareaClose Out: %s\n", out)
	return out
}

// RemoveTagsUntilTextareaClose removes all the tags before the closing textarea one
func RemoveTagsUntilTextareaClose(src string) (out string) {
	re := regexp.MustCompile(`(?i)(<([^>]+)>)`)
	m := re.FindAllStringIndex(src, -1)
	last := 0
	// revisit once https://github.com/golang/go/issues/5690 is resolved
	for i := 0; i < len(m); i++ {
		// copy the substring from "last" until the current match
		out += src[last:m[i][0]]
		if strings.HasPrefix(strings.ToLower(src[m[i][0]:m[i][1]]), "</textarea>") {
			out += src[m[i][0]:]
			return out
		}
		// otherwise, skip the tag
		last = m[i][1]
		// leave for debugging
		// fmt.Printf("%d %s\n", i, src[m[i][0]:m[i][1]])
	}
	if last < len(src) {
		out += src[last:]
	}
	// leave for debugging
	// fmt.Printf("RemoveTagsUntilTextareaClose Out: %s\n", out)
	return out
}

// ReplaceTextareaSafe removes all the tags after the closing textarea one
func ReplaceTextareaSafe(src string) (out string) {
	re := regexp.MustCompile(`(?i)(<([^>]+)>)`)
	m := re.FindAllStringIndex(src, -1)
	last := 0
	// revisit once https://github.com/golang/go/issues/5690 is resolved
	for i := 0; i < len(m); i++ {
		// copy the string until the current match
		out += src[last:m[i][0]]
		if strings.HasPrefix(strings.ToLower(src[m[i][0]:m[i][1]]), "</textarea>") {
			out += src[m[i][0]:m[i][1]]
			out += Transform(src[m[i][1]:], TagsOff)
			last = len(src)
			break
		} else {
			// copy verbatim before </textarea>
			out += src[m[i][0]:m[i][1]]
			last = m[i][1]
		}
		// leave for debugging
		// fmt.Printf("%d %s\n", i, src[m[i][0]:m[i][1]])
	}
	if last < len(src) {
		out += src[last:]
	}
	// leave for debugging
	// fmt.Printf("TextareaSafe Out: %s\n", out)
	return out
}

func removeTextareaClose(in string) (out string) {
	re := regexp.MustCompile(`(?i)(</textarea\s*>)`)
	out = re.ReplaceAllLiteralString(in, "")
	return
}

func backslashDoublequotes(in string) (out string) {
	for _, r := range in {
		switch r {
		case '"':
			out += `\"`
		case '\\':
			out += `\\`
		default:
			out += string(r)
		}
	}
	return
}

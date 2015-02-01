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

// File has two parts - one related to the map of filter fields to
// the replacers (for more standard replacements).
// The second part are the functions that do transformations
// beyond simple string substitution (regexp etc.)

type transformerMap map[filter]*strings.Replacer

var trMap transformerMap

// initialize the main replacing functions
func init() {
	trMap = transformerMap(make(map[filter]*strings.Replacer))
	trMap[QuotesOff] = strings.NewReplacer(`"`, "", `'`, "")
	trMap[QuotesCook] = strings.NewReplacer(`"`, `&quot;`, `'`, `&#39;`)
	trMap[DoubleQuotesOff] = strings.NewReplacer(`"`, "")
	trMap[DoubleQuotesCook] = strings.NewReplacer(`"`, `&quot;`)
	trMap[SingleQuotesOff] = strings.NewReplacer(`'`, "")
	trMap[SingleQuotesCook] = strings.NewReplacer(`'`, `&#39;`)
	trMap[BackslashEscape] = strings.NewReplacer(`\`, `\\`)
	trMap[DoubleQuotesBackslashEscape] = strings.NewReplacer(`"`, `\"`)
	trMap[TagsOff] = strings.NewReplacer(`<`, "", `>`, "")
	trMap[TagsCook] = strings.NewReplacer(`<`, `&lt;`, `>`, `&gt;`)
	trMap[GreaterThanOff] = strings.NewReplacer(`>`, "")
	trMap[GreaterThanCook] = strings.NewReplacer(`>`, `&gt;`)
	trMap[LessThanOff] = strings.NewReplacer(`<`, "")
	trMap[LessThanCook] = strings.NewReplacer(`<`, `&lt;`)
	trMap[SpacesOff] = strings.NewReplacer(` `, "")
	trMap[SpacesCook] = strings.NewReplacer(` `, "&#20;")
	trMap[ParensOff] = strings.NewReplacer(`(`, "", `)`, "")
}

// Transformer transforms a string by escaping, filtering or other modification
type Transformer interface {
	Transform(s string) string
}

type Filter struct {
	options []filter
	nested  *Transformer
}

func (f Filter) Transform(s string) string {
	if f.nested != nil {
		s = (*f.nested).Transform(s)
	}
	for _, opt := range f.options {
		if tr, ok := trMap[opt]; ok {
			s = tr.Replace(s)
		} else {
			log.Printf("ERROR in Transform - option %v is not in the trMap! Skipping.")
		}
	}
	return s
}

type RegexpTransformer struct {
	re     *regexp.Regexp
	nested *Transformer
}

// RegexpTransformer transforms a string by applying its regular expression
func (r RegexpTransformer) Transform(s string) string {
	if r.nested != nil {
		s = (*r.nested).Transform(s)
	}

	return r.re.ReplaceAllLiteralString(s, "")
}

func NewTransformerRegexp(s string) Transformer {
	re := regexp.MustCompile(s)
	return RegexpTransformer{re, nil}
}

// NewTransformer takes a (possibly empty) list of filter options
// and returns a corresponding Transformer
func NewTransformer(fltr ...filter) Transformer {
	f := Filter{}
	f.options = make([]filter, 0)
	for _, filter := range fltr {
		f.options = append(f.options, filter)
	}
	return f
}

// StripScript removes opening and closing script tags
func StripScript(s string) string {
	tr := NewTransformerRegexp(`</?script[^>]*>`)
	return tr.Transform(s)
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
	return string(i)
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

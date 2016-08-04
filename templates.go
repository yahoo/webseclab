// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	ht "html/template"
	tt "text/template"
)

//go:generate go run genstatic.go

func init() {
	parseTemplates()
}

// TemplateData wraps and embeds data related to the template processing
type TemplateData struct {
	path  string
	fixed bool
	InData
}

// LookupTextTemplate returns a pointer to the parsed text template and true if lookup successful,
// or nil and false if no template for the given name was found
func LookupTextTemplate(name string) (ttmpl *tt.Template, ok bool) {
	ttmpl, ok = _ctx.tmplsT[name]
	if !ok {
		return nil, false
	}
	return ttmpl, ok
}

// AddTextTemplate adds template record for the given string
func AddTextTemplate(name string, ttmpl *tt.Template) {
	_ctx.tmplsT[name] = ttmpl
}

func textTemplates() map[string]*tt.Template {
	return _ctx.tmplsT
}

// LookupHTMLTemplate returns a pointer to the parsed template and true if lookup successful,
// or nil and false if no template for the given name was found
func LookupHTMLTemplate(name string) (htmpl *ht.Template, ok bool) {
	htmpl, ok = _ctx.tmplsH[name]
	if !ok {
		return nil, false
	}
	return htmpl, ok
}

// AddHTMLTemplate adds template record for the given string
func AddHTMLTemplate(name string, htmpl *ht.Template) {
	_ctx.tmplsH[name] = htmpl
}

func htmlTemplates() map[string]*ht.Template {
	return _ctx.tmplsH
}

// ParseTemplates parses the templates from the generated file static.go
func parseTemplates() error {
	header := Templates["common/header"]
	footer := Templates["common/footer"]
	for path, templstr := range Templates {
		tTmpl := tt.New(path)
		hTmpl := ht.New(path)
		tTmpl = tt.Must(tTmpl.Parse(templstr))
		tt.Must(tTmpl.Parse(header))
		tt.Must(tTmpl.Parse(footer))
		AddTextTemplate(path, tTmpl)
		hTmpl = ht.Must(hTmpl.Parse(templstr))
		ht.Must(hTmpl.Parse(header))
		ht.Must(hTmpl.Parse(footer))
		AddHTMLTemplate(path, hTmpl)
	}
	return nil
}

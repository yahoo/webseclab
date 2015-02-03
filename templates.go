// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"errors"
	ht "html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	tt "text/template"
)

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
	if ok == false {
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

// LookupHtmlTemplate returns a pointer to the parsed template and true if lookup successful,
// or nil and false if no template for the given name was found
func LookupHtmlTemplate(name string) (htmpl *ht.Template, ok bool) {
	htmpl, ok = _ctx.tmplsH[name]
	if ok == false {
		return nil, false
	}
	return htmpl, ok
}

// AddHtmlTemplate adds template record for the given string
func AddHtmlTemplate(name string, htmpl *ht.Template) {
	_ctx.tmplsH[name] = htmpl
}

func htmlTemplates() map[string]*ht.Template {
	return _ctx.tmplsH
}

// ParseTemplates parses the templates found in the base directory
func ParseTemplates(base string, skiphtml ...string) error {
	if base == "" {
		log.Fatalf("ERROR in ParseTemplates - base parameter is empty!")
		return errors.New("ERROR in ParseTemplates - base parameter is empty!")
	}
	if base[len(base)-1] != '/' {
		base += "/"
	}

	tmplfiles, err := getTemplateFiles(base)
	if err != nil {
		return err
	}
	if len(tmplfiles) == 0 {
		return errors.New("No template files found under " + base)
	}

	for _, filename := range tmplfiles {
		htmpl := ht.New(filename)
		ttmpl := tt.New(filename)
		ttmpl = tt.Must(ttmpl.ParseFiles(base+filename, base+"/common/header", base+"/common/footer"))
		AddTextTemplate(filename, ttmpl)
		htmpl = ht.Must(htmpl.ParseFiles(base+filename, base+"/common/header", base+"/common/footer"))
		AddHtmlTemplate(filename, htmpl)
	}
	return nil
}

// getTemplateFiles collects all files under the base dir
func getTemplateFiles(base string) (files []string, err error) {
	files = make([]string, 0, 50)
	visit := func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return errors.New("File " + path + " does not exist or not readable")
		}
		// don't add common shared files
		if strings.HasSuffix(path, "common/footer") ||
			strings.HasSuffix(path, "common/header") {
			return nil
		}
		if f.IsDir() == false {
			files = append(files, strings.TrimPrefix(path, base))
		}
		return err
	}
	err = filepath.Walk(base, visit)
	return files, err
}

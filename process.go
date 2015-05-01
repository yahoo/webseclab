// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func DoLabTestStandard(w http.ResponseWriter, r *http.Request) *LabResp {
	// split r.URL.RawQuery into a param map
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	input := &InData{}
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
		unesc, err := url.QueryUnescape(input.InRaw)
		if err != nil {
			fmt.Printf("ERROR in url.QueryUnescape on %s\n", input.InRaw)
		}
		input.In = unesc
	}
	err := DoTemplate(w, r.URL.Path, input)
	if err != nil {
		httpcode := http.StatusInternalServerError
		log.Printf("Error returned from DoTemplate: %s\n", err)
		if _, ok := err.(ErrNotFound); ok {
			httpcode = http.StatusNotFound
		}
		return &LabResp{Err: err, Code: httpcode}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

type ErrNotFound struct {
	text string
}

func NewErrNotFound(text string) (e ErrNotFound) {
	e = ErrNotFound{text: text}
	return
}

func (e ErrNotFound) Error() string {
	return e.text
}

// DoTemplate opens the template file and processes the template with the passed input
// if the URL Path ends with ".ok", it uses an HTML context-escaped template
// (html/template - safe version), otherwise - text/template (exploitable)
func DoTemplate(w http.ResponseWriter, path string, input *InData) (err error) {
	if input == nil {
		return errors.New("ERROR Internal - Nil passed to DoTemplate as InData")
	}
	if path == "" {
		return errors.New("ERROR Internal - empty path passed to DoTemplate")
	}
	if len(path) == 1 {
		return errors.New("ERROR Internal - too short (len = 1) path passed to DoTemplate: " + path)
	}
	hasOk := strings.HasSuffix(path, ".ok")
	var fpath string // filepath
	if hasOk {
		fpath = path[1 : len(path)-3]
	} else {
		fpath = path[1:]
	}
	if hasOk {
		// try text template with the literal ".ok" first for special cases (templates with broken HTML)
		_, ok := LookupTextTemplate(fpath + ".ok")
		// text/template - no context-sensitive escaping
		if ok {
			return doTextTemplate(w, fpath+".ok", input)
		}
		return doHtmlTemplate(w, fpath, input)
	}
	return doTextTemplate(w, fpath, input)
}

func doHtmlTemplate(w http.ResponseWriter, fpath string, input *InData) (err error) {
	// html/template - context-sensitive escaping
	tmpl, ok := LookupHtmlTemplate(fpath)
	if ok == false {
		return errors.New("Error in DoTemplate - html template " + fpath + " not found.")
	}
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	parts := strings.Split(fpath, "/")
	err = tmpl.ExecuteTemplate(w, parts[len(parts)-1], *input)
	if err != nil {
		log.Printf("Error in DoTemplate (html) - tmpl.Execute: %s\n", err)
		return err
	}
	return nil
}

func doTextTemplate(w http.ResponseWriter, fpath string, input *InData) (err error) {
	tmpl, ok := LookupTextTemplate(fpath)
	if ok == false {
		err := NewErrNotFound("Error in DoTemplate - text template " + fpath + " not found.")
		return err
	}
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	parts := strings.Split(fpath, "/")
	err = tmpl.ExecuteTemplate(w, parts[len(parts)-1], *input)
	if err != nil {
		log.Printf("Error in DoTemplate (text) - tmpl.Execute's err: %s\n", err)
		return err
	}
	return nil
}

// HandleFilterBased factors out the common handling for the "typical" -
// filter-based tests (listed in the map in custom.go, filters are in filters.go)
func HandleFilterBased(w http.ResponseWriter, r *http.Request, filters []filter) *LabResp {
	var input InData
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
		unesc, err := url.QueryUnescape(input.InRaw)
		if err != nil {
			log.Printf("Error in %s: %s\n", r.URL.Path, err)
			return &LabResp{Err: nil, Code: http.StatusInternalServerError}
		}
		input.In = Transform(unesc, filters...)
	}
	err := DoTemplate(w, r.URL.Path, &input)
	if err != nil {
		log.Printf("Error in DoTemplate: %s\n", err)
		return &LabResp{Err: nil, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

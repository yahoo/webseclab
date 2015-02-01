// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// LabResp wraps the data related to processing results
type LabResp struct {
	Err      error
	Code     int
	Redirect string
	Path     string
	Fixed    bool
	InData
}

// Strings converts LabResp to a string
func (r *LabResp) String() string {
	return "Error: " + r.Err.Error() + ", code: " + strconv.Itoa(r.Code) + ", redirect: " + r.Redirect
}

// LabHandler is the main http handler type
type LabHandler func(http.ResponseWriter, *http.Request) *LabResp

// ServeHTTP implements the interface required by http.Handle
func (fn LabHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := fn(w, r)
	if resp.Code == http.StatusFound {
		http.Redirect(w, r, resp.Redirect, http.StatusFound)
		return
	}
	if resp.Err != nil && resp.Code != 0 {
		w.WriteHeader(resp.Code)
		w.Header().Set("Content-type", "text/html; charset=utf-8")
		log.Printf("ERROR in LabHandler, errorText: %s\n", resp.Err.Error())
		tmpl, err := template.New("Error").Parse("ERROR: {{.}}\n")
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(w, resp.Err.Error())
		return
	}
}

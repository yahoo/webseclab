// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
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

// MakeMainHandler performs routing between 'standard' (template based with standard parameters) cases
// and those requiring custom processing.  For standard processing, it unescapes input and prepars an instance of Indata
// which is then passed to the template execution.  For URLs found in the map of custom processing,
// the corresponding function is called.
func MakeMainHandler(noindex bool) LabHandler {
	return func(w http.ResponseWriter, r *http.Request) *LabResp {
		// routing - handle a special case, path = "/" (=> index.html)
		// dump, _ := httputil.DumpRequest(r, true)
		// fmt.Printf("DEBUG request dump: %q\n", dump)
		indexFn := MakeIndexFunc("index.html")
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			if noindex {
				return &LabResp{
					Err:  errors.New("index page is prevented with -noindex option"),
					Code: http.StatusForbidden,
				}

			}
			indexFn(w, r)
			return &LabResp{Err: nil, Code: http.StatusOK}
		}
		// be paranoid about the non-IP domains to protect cookies against XSS
		safe := IsSafeHost(r.Host)
		if !safe {
			ipurl, err := GetIPURL(r.Host, r.URL)
			if err != nil {
				return &LabResp{Err: errors.New("ERROR in GetIPUrl(" + r.URL.String() + "): " + err.Error()),
					Code: http.StatusInternalServerError}
			}
			return &LabResp{Err: errors.New(r.Host + " - not an IP quad pair"),
				Code:     http.StatusFound,
				Redirect: ipurl.String()}
		}

		// check if custom handling is needed
		funcmap := CustomMap()
		handler, ok := funcmap[r.URL.Path]
		if ok {
			return handler(w, r)
		}
		filterMap := filterMap()
		filters, ok := filterMap[r.URL.Path]
		if ok {
			return HandleFilterBased(w, r, filters)
		}
		return DoLabTestStandard(w, r)
	}
}

// Ruok replies with an ack to a ping:
// "ruok" => "imok\n" (for /ruok monitoring entrypoint).
func Ruok(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "imok\n")
}

// MakeExitFunc creates an /exit handler.
func MakeExitFunc(ln net.Listener) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.RemoteAddr, ":")
		if len(parts) > 0 {
			if parts[0] != "127.0.0.1" && parts[0] != "localhost" {
				w.WriteHeader(http.StatusForbidden)
				io.WriteString(w, "Access denied. (/exit called from non-local IP: "+parts[0]+")\n")
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, `bye`)
		log.Println("Received exit request, exiting")
		ln.Close()
	}
}

// MakeIndexFunc creates a function to display the index file (ToC)
func MakeIndexFunc(page string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Request for %s from %s\n", r.URL.String(), r.RemoteAddr)
		var data = &InData{}
		// for UI, find out an Ip quad-pair link if we are not on a such already
		// this is done to protect cookies of our domain against XSS (see ip.go)
		if !IsIP(r.Host) {
			iplink, err := GetIPURL(r.Host, r.URL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`Internal Server Error`))
				log.Printf("ERROR - unable to find out my own IP address: r.Host = %s, r.URL = %s\n", r.Host, r.URL.String())
				return
			}
			data.In = iplink.Host
		}
		err := DoTemplate(w, "/index.html", data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`Internal Server Error`))
			return
		}
		return
	}
}

// MakeStaticFunc creates a "static function" processing a template
// with empty data input.
func MakeStaticFunc() LabHandler {
	return func(w http.ResponseWriter, r *http.Request) *LabResp {
		err := DoTemplate(w, r.URL.Path, &InData{})
		if err != nil {
			log.Printf("Error in MakeStaticFunc, DoTemplate returned Err: %s\n", err)
			return &LabResp{Err: err, Code: http.StatusInternalServerError}
		}
		return &LabResp{Err: nil, Code: http.StatusOK}
	}
}

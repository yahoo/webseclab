// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// KillPredecessor check if there is a listener on the given port,
// and if so, sends it a command to exit.
// This allows to start a new copy (post-build etc.) with no errors
// and still have only one instance of webseclab running.
func KillPredecessor(port string) {
	const pauseT = 500
	if port == "" {
		log.Fatal("ERROR in main.killPredecessor - port string is empty!")
		return
	}
	res, err := http.Get("http://127.0.0.1" + port + "/exit")
	if err == nil {
		defer res.Body.Close()
		d, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Error reading the response from %s: %s\n", port, err)
			return
		}
		log.Printf("(INFO) killPredecessor - predecessor response to our termination request: %s\n", string(d))
		// give time to OS to make the port available
		time.Sleep(pauseT * time.Millisecond)
	}
}

// ack to a ping: "ruok" => "imok\n" (for /ruok monitoring entrypoint)
func Ruok(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "imok\n")
}

// MakeExitFunc creates an /exit handler
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
func MakeIndexFunc(base, page string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Request for %s from %s\n", r.URL.String(), r.RemoteAddr)
		var data = &InData{}
		// for UI, find out an Ip quad-pair link if we are not on a such already
		// this is done to protect cookies of our domain against XSS (see ip.go)
		if IsIp(r.Host) == false {
			iplink, err := GetIpUrl(r.Host, r.URL)
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

// MakeMainHandler performs routing between 'standard' (template based with standard parameters) cases
// and those requiring custom processing.  For standard processing, it unescapes input and prepars an instance of Indata
// which is then passed to the template execution.  For URLs found in the map of custom processing,
// the corresponding function is called.
func MakeMainHandler(base string) LabHandler {
	return func(w http.ResponseWriter, r *http.Request) *LabResp {
		// routing - handle a special case, path = "/" (=> index.html)
		// dump, _ := httputil.DumpRequest(r, true)
		// fmt.Printf("DEBUG request dump: %q\n", dump)
		indexFn := MakeIndexFunc(base, "index.html")
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			indexFn(w, r)
			return &LabResp{Err: nil, Code: http.StatusOK}
		}
		// be paranoid about the non-IP domains to protect cookies against XSS
		safe := IsSafeHost(r.Host)
		if safe == false {
			ipurl, err := GetIpUrl(r.Host, r.URL)
			if err != nil {
				return &LabResp{Err: errors.New("ERROR in GetIpUrl(" + r.URL.String() + "): " + err.Error()),
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
		filtermap := FilterMap()
		filters, ok := filtermap[r.URL.Path]
		if ok {
			return HandleFilterBased(w, r, filters)
		}
		// check the prefix cases as well - HACK!! [TODO: find a more elegant way]

		return DoLabTestStandard(w, r)
	}
}

// MakeStaticFunc
func MakeStaticFunc(base string) LabHandler {
	return func(w http.ResponseWriter, r *http.Request) *LabResp {
		if len(base) > 0 && base[len(base)-1] != '/' {
			base += "/"
		}
		err := DoTemplate(w, r.URL.Path, &InData{})
		if err != nil {
			log.Printf("Error in MakeStaticFunc, DoTemplate returned Err: %s\n", err)
			return &LabResp{Err: err, Code: http.StatusInternalServerError}
		}
		return &LabResp{Err: nil, Code: http.StatusOK}
	}
}

// CheckPath checks if the given file path exists
func CheckPath(path string) (err error) {
	if _, err = os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

// TemplateBaseDefault returns the default value for the default base directory
// It is the templates subdirectory of the current working directory,
// $GOPATH/src/github.com/yahoo/webselab/templates if $GOPATH is set
// and the webseclab directory is present
// or an empty string
func TemplateBaseDefault() (s string, err error) {
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Printf("Unable to get the current directory in TemplateBaseDefault, returning empty string: %s\n", err)
		return "", err
	}
	if CheckPath(path.Join(pwd, "templates")) == nil {
		return path.Join(pwd, "templates"), nil
	}
	gopath := os.Getenv("GOPATH")

	if gopath == "" {
		return "", errors.New("No GOPATH in the environment in TemplateBaseDefault, bailing out")
	}
	// fix-up for the case (as with Travis) that GOPATH ends with the ':'
	// - colon separator
	if strings.HasSuffix(gopath, ":") {
		gopath = gopath[:len(gopath)-1]
	}
	base := path.Join(gopath, "src/github.com/yahoo/webseclab/templates")
	err = CheckPath(base)
	if err == nil {
		return base, nil
	}
	return "", err
}

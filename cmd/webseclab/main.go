// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// server for Webseclab - set of web application security test
// WARNING - the pages are intentionally insecure! Be careful!
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/yahoo/webseclab"
)

// type indexHandler func(http.ResponseWriter, *http.Request) error

// func (fn indexHandler) ServerHTTP(w http.ResponseWriter, r *http.Request) {
// 	fn(w, r)
// }

const notice = `Attention: Webseclab is purposedly INSECURE software intended for testing and education.  Use it at your own risk and be careful! Hit Ctrl-C now if you don't understand the risks invovled.

Webseclab executable includes the Go runtime which is covered by the Go license that can be found in http://golang.org/LICENSE.  See https://github.com/yahoo/webseclab for the Webseclab source code, license, and additional information.
`

func main() {
	// use up to 2 CPUs, not more
	cpus := runtime.NumCPU()
	if cpus >= 2 {
		cpus = 2
	}
	runtime.GOMAXPROCS(cpus)

	port := flag.String("http", ":8080", "port to run the webserver on")
	noindex := flag.Bool("noindex", false, "do not serve the top index page (/ and /index.html)")
	cleanup := flag.Bool("cleanup", false, "cleanup only (terminate existing instance and exit)")
	version := flag.Bool("version", false, "display version number and exit")
	flag.Parse()
	if *version {
		fmt.Println(webseclab.WebseclabVersion)
		os.Exit(0)
	}
	fmt.Println(notice)
	// if there is another webseclab server running, tell it to quit and make way for us
	webseclab.KillPredecessor(*port)
	if *cleanup {
		return
	}
	ln, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Webseclab starts to listen on %s\n", *port)

	if !*noindex {
		http.HandleFunc("/index.html", webseclab.MakeIndexFunc("/index.html"))
	}
	http.HandleFunc("favicon.ico", func(w http.ResponseWriter, r *http.Request) {
	})
	http.Handle("/", webseclab.MakeMainHandler(*noindex))
	http.HandleFunc("/exit", webseclab.MakeExitFunc(ln))
	http.HandleFunc("/ruok", webseclab.Ruok)

	http.Serve(ln, nil)
}

// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// server for Webseclab - set of web application security test
// WARNING - the pages are intentionally insecure! Be careful!
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/yahoo/webseclab"
)

type indexHandler func(http.ResponseWriter, *http.Request) error

func (fn indexHandler) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	fn(w, r)
	return
}

func main() {
	// use up to 2 CPUs, not more
	cpus := runtime.NumCPU()
	if cpus >= 2 {
		cpus = 2
	}
	runtime.GOMAXPROCS(cpus)
	defaultBase, err := webseclab.TemplateBaseDefault()
	if err != nil {
		panic(err)
	}
	base := flag.String("base", defaultBase, "base path for webseclab templates")
	port := flag.String("http", ":8080", "port to run the webserver on")
	noindex := flag.Bool("noindex", false, "do not serve the top index page (/ and /index.html)")
	cleanup := flag.Bool("cleanup", false, "cleanup only (terminate existing instance and exit)")
	flag.Parse()
	fmt.Printf("Using %s as the template base, change it with -base command-line option (run 'webseclab -help' for usage help)\n", *base)
	var abspath string
	if path.IsAbs(*base) {
		abspath = *base
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		abspath = path.Join(pwd, *base)
	}
	if err := webseclab.CheckPath(abspath); err != nil {
		log.Printf("Unable to open base %s: %s\n", abspath, err)
		os.Exit(1)
	}
	// if there is another webseclab server running, tell it to quit and make way for us
	webseclab.KillPredecessor(*port)
	if *cleanup {
		return
	}
	err = webseclab.ParseTemplates(abspath)
	if err != nil {
		panic(err)
	}
	ln, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Webseclab starts to listen on %s\n", *port)

	if !*noindex {
		http.HandleFunc("/index.html", webseclab.MakeIndexFunc(*base, "/index.html"))
	}
	http.HandleFunc("favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		return
	})
	http.Handle("/", webseclab.MakeMainHandler(*base, *noindex))
	http.HandleFunc("/exit", webseclab.MakeExitFunc(ln))
	http.HandleFunc("/ruok", webseclab.Ruok)

	http.Serve(ln, nil)
}

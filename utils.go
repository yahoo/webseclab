// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"io/ioutil"
	"log"
	"net/http"
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

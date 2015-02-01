// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"net/http"
	"strings"
)

// ParseRawQuery copied from net/url parseQuery but without unescaping keys/values
func ParseRawQuery(m map[string][]string, query string) {
	for query != "" {
		key := query
		if i := strings.Index(key, "&"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		m[key] = append(m[key], value)
	}
	return
}

// Input extracts the escaped and "raw" values of in parameters
func Input(r *http.Request) *InData {
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	input := InData{In: r.FormValue("in")}
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
	}
	return &input
}

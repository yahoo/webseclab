// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import "net/http"

// mockResponseWriter is a mock substitute for http://golang.org/pkg/net/http/#ResponseWriter
type mockResponseWriter struct {
	header http.Header
}

func newMockResponseWriter() mockResponseWriter {
	m := mockResponseWriter{}
	m.header = make(http.Header)
	return m
}

func (r mockResponseWriter) Header() (header http.Header) {
	return r.header
}

func (r mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (r mockResponseWriter) WriteHeader(int) {
}

// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import "net/http"

type MockResponseWriter struct {
	header http.Header
}

func NewMockResponseWriter() MockResponseWriter {
	m := MockResponseWriter{}
	m.header = make(http.Header)
	return m
}

func (r MockResponseWriter) Header() (header http.Header) {
	return r.header
}

func (r MockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (r MockResponseWriter) WriteHeader(int) {
	return
}

// func TestXssDoubq1(t *testing.T) {
// 	t.Parallel()
// 	var w = NewMockResponseWriter()
// 	path := "/a/b/c"
// 	req, err := http.NewRequest("GET", path, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	resp := XssDoubq1(w, req)
// 	if resp.Err != nil {
// 		t.Error(resp.Err)
// 	}
// }

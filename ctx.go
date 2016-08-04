// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	ht "html/template"
	tt "text/template"
)

type ctx struct {
	tmplsH map[string]*ht.Template
	tmplsT map[string]*tt.Template
}

// package global
var _ctx ctx

func init() {
	_ctx.tmplsH = make(map[string]*ht.Template)
	_ctx.tmplsT = make(map[string]*tt.Template)
}

// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckPathFake(t *testing.T) {
	t.Parallel()
	fakepath := `/tmp/foo/bar/no/such/path`
	if CheckPath(fakepath) == nil {
		t.Errorf("Expecting non-nil error on the fakepath %s\n", fakepath)
	}
}

func TestCheckPathReal(t *testing.T) {
	t.Parallel()
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		t.Errorf("Unable to get the current directory: %s\n", err)
		return
	}
	if CheckPath(pwd) != nil {
		t.Errorf("Expecting nil error on the current directory path %s\n", pwd)
	}
}

func TestTemplateBaseDefault(t *testing.T) {
	t.Parallel()
	base, err := TemplateBaseDefault()
	if err != nil {
		t.Errorf("Error in TemplateBaseDefault call: %s\n", err)
		return
	}
	gopath := os.Getenv("GOPATH")
	// fix-up for the case (as with Travis) that GOPATH ends with the ':'
	// - colon separator
	if strings.HasSuffix(gopath, ":") {
		gopath = gopath[0 : len(gopath)-1]
	}
	want := path.Join(gopath, "src/github.com/yahoo/webseclab/templates")
	if base != want {
		t.Errorf("Want template base %s, got %s\n", want, base)
	}
}

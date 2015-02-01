// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"os"
	"testing"
)

func TestParseTemplates(t *testing.T) {
	t.Parallel()
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := ParseTemplates(pwd + "/templates"); err != nil {
		t.Errorf("Error in ParseTemplates (text: %s\n", err)
		return
	}
}

// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import "testing"

func TestParseTemplates(t *testing.T) {
	t.Parallel()
	if err := parseTemplates(); err != nil {
		t.Errorf("Error in ParseTemplates (text: %s\n", err)
		return
	}
}

// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import "testing"

func TestParseRawQuery(t *testing.T) {
	t.Parallel()
	query := "a=01&b=02&c=<xss1>"
	m := make(map[string][]string)
	ParseRawQuery(m, query)
	// fmt.Printf("Result: %v\n", m)
	if _, ok := m["b"]; ok == false {
		t.Error("No key/value pair for 'b' - want 'b' => 02'")
	}
}

func TestParseRawQueryMultValues(t *testing.T) {
	t.Parallel()
	query := "a=01&b=02&b=02a&b=02c"
	m := make(map[string][]string)
	ParseRawQuery(m, query)
	// fmt.Printf("Result: %v\n", m)
	if el, ok := m["b"]; ok == false {
		t.Error("No key/value pair for 'b' - want 'b' => 02'")
		return
	} else {
		if len(el) != 3 {
			t.Errorf("Wrong value of elements for key 'b' - expected 3, got: %d\n", len(el))
			return
		}
		if el[2] != "02c" {
			t.Errorf("Unexpected value for the 3rd value of key 'b' - want 02c, got %s\n", el[2])
		}
	}
}

func TestParseRawQueryClosingScriptTag(t *testing.T) {
	t.Parallel()
	inj := `"><script>alert("xss");</script>`
	query := "a=01&b=02&c=" + inj
	m := make(map[string][]string)
	ParseRawQuery(m, query)
	// fmt.Printf("Result: %v\n", m)
	if _, ok := m["c"]; ok == false {
		t.Error("No key/value pair for 'c' - want 'c' => " + inj)
		return
	}
	if len(m["c"]) != 1 {
		t.Errorf("Want 1 value of m[\"c\"], got: %d\n", len(m["c"]))
	}
	if m["c"][0] != inj {
		t.Errorf("Wrong value - want %s, got %s\n", inj, m["c"])
	}
}

// This test makes sure that your build pipeline properly reacts on test failures
//
// func TestBadDummy(t *testing.T) {
// 	t.Parallel()
// 	if true {
// 		t.Errorf("Dummy Test Failure")
// 	}
// }

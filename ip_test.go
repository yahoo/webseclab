// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

import (
	"net/url"
	"regexp"
	"testing"
)

func TestIsSafeHost(t *testing.T) {
	t.Parallel()
	if !IsSafeHost("localhost") {
		t.Error("IsSafeHost must return true for 'localhost', got false")
	}
	if !IsSafeHost("localhost:8088") {
		t.Error("IsSafeHost must return true for 'localhost', got false")
	}
	if !IsSafeHost("127.0.0.1") {
		t.Error("IsSafeHost must return true for '127.0.0.1', got false")
	}
	if !IsSafeHost("example.com") {
		t.Error("IsSafeHost must return false for a FQDN , got true")
	}
}

func TestIPRegexp(t *testing.T) {
	t.Parallel()
	var validIP = regexp.MustCompile(`^([\d]{1,3}\.){3}[\d]{1,3}(:\d+)?$`)
	// var validIP = regexp.MustCompile(`^[:digit]+\.[:digit]+\.[:digit]+`)
	if !validIP.MatchString("127.0.0.1") {
		t.Errorf("Regexp did not match quad-pair IP string with no port")
	}
	if !validIP.MatchString("127.0.0.1:8080") {
		t.Errorf("Regexp did not match quad-pair IP string with port")
	}
}

func TestIsIP(t *testing.T) {
	t.Parallel()
	if !IsIP("127.0.0.1") {
		t.Errorf("Regexp did not match quad-pair IP string with no port")
	}
	if IsIP("www.example.com") {
		t.Errorf("Regexp matched domain that is not a quad-pair IP")
	}
}

func TestIsIPURL(t *testing.T) {
	t.Parallel()
	var table = []struct {
		in   string
		want bool
	}{
		{"http://subdomain.example.com:8080/a/b/c/d?in=fake", false},
		{"http://subdomain.example.com/a/b/c/d?in=fake", false},
		{"http://10.89.01.02:8080/a/b/c/d?in=fake", true},
		{"http://10.89.03.04/a/b/c/d?in=fake", true},
	}
	for _, i := range table {
		u, err := url.Parse(i.in)
		if err != nil {
			t.Errorf("Error parsing URL %s: %s\n", i.in, err)
			return
		}
		if i.want != IsIPURL(u) {
			t.Errorf("Wrong result in IsIPURL on %s: want %t\n", u.String(), i.want)
		}
	}
}

// func TestGetIpUrl(t *testing.T) {
// 	t.Parallel()
// 	s := "http://webseclab.yahoo-inc.com"
// 	want := "http://10.89.01.02"
// 	u, err := url.Parse(s)
// 	if err != nil {
// 		t.Errorf("Error parsing URL %s: %s\n", s, err)
// 		return
// 	}
// 	ipurl, err := GetIPUrl(u.Host, u)
// 	if err != nil {
// 		t.Errorf("Error getting IP quad URL for %s: %s\n", s, err)
// 		return
// 	}
// 	if ipurl.String() != want {
// 		t.Errorf("Want %s, got %s\n", want, ipurl.String())
// 	}
// }

// func TestGetIPUrlWithPort(t *testing.T) {
// 	t.Parallel()
// 	s := "http://example.yahoo.com:8088"
// 	want := "http://93.184.216.34:8088"
// 	u, err := url.Parse(s)
// 	if err != nil {
// 		t.Errorf("Error parsing URL %s: %s\n", s, err)
// 		return
// 	}
// 	ipurl, err := GetIPUrl(u.Host, u)
// 	if err != nil {
// 		t.Errorf("Error getting IP quad URL for %s: %s\n", s, err)
// 		return
// 	}
// 	if ipurl.String() != want {
// 		t.Errorf("Want %s, got %s\n", want, ipurl.String())
// 	}
// }

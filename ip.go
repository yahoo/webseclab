// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webseclab

// When running intentionally XSS-vulnerable site in a valuable domain
// such as examplecompany.com, we want to protect the cookies
// therefore redirect all requests to the IP addresses (quad pairs)

import (
	"errors"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// IsSafe check if the host is an IP quad pair or localhost
func IsSafeHost(s string) bool {
	var domain string
	if strings.ContainsRune(s, ':') {
		parts := strings.Split(s, ":")
		domain = parts[0]
	} else {
		domain = s
	}
	if domain == "localhost" {
		return true
	}
	return IsIp(s)
}

// IsIP checks if the argument is a IP quad pair such as 101.02.03.04 (with optional port ex. :8080)
func IsIp(s string) bool {
	p := regexp.MustCompile(`^([\d]{1,3}\.){3}[\d]{1,3}(:\d+)?$`)
	return p.MatchString(s)
}

// isIPUrl checks if the URL is a IP quad pair such as 101.02.03.04 (with optional port ex. :8080)
func IsIpUrl(u *url.URL) bool {
	return IsIp(u.Host)
}

// GetIpURL returns a corresponding IP-quad URL if a FQDN is used
// if there are multiple results from LookupHost, the first one is returned
func GetIpUrl(host string, link *url.URL) (*url.URL, error) {
	if IsIpUrl(link) {
		return link, nil
	}
	var domain, port string
	if host == "" && link.Host == "" {
		return link, errors.New("Error in GetIpUrl - no host available (neither in host param nor inside of Url). Host = " + host + ", url = " + link.String())
	}
	if host != "" {
		link.Host = host
	}
	if strings.ContainsRune(link.Host, ':') {
		parts := strings.Split(link.Host, ":")
		domain = parts[0]
		port = parts[1]
		if len(port) == 0 {
			domain = link.Host
		}
	} else {
		domain = link.Host
	}
	ipquads, err := net.LookupHost(domain)
	if err != nil {
		log.Printf("ERROR in IpCheck - unable to lookup the IP of %s, error: %s\n", link.Host, err)
		return link, errors.New("Internal error - DNS lookup unavailable")
	}
	if len(ipquads) == 0 {
		log.Printf("ERROR in IpCheck - unable to lookup the IP of %s, error: %s\n", link.Host, err)
		return link, errors.New("Internal error - DNS lookup unavailable")
	}
	if len(port) == 0 {
		link.Host = ipquads[0]
	} else {
		// special case for localhost testing
		if ipquads[0] != "::1" {
			link.Host = ipquads[0] + ":" + port
		} else {
			link.Host = ipquads[1] + ":" + port
		}
	}
	if link.Scheme == "" {
		link.Scheme = "http"
	}
	return link, nil
}

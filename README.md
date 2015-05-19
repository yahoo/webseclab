### Webseclab [![Build Status](https://travis-ci.org/yahoo/webseclab.svg?branch=master)](https://travis-ci.org/yahoo/webseclab)

Webseclab contains a sample set of web security test cases and a toolkit to construct new ones.  It can be used for testing security scanners, to replicate or reconstruct issues, or to help with investigations or discussions of particular types of web security bugs.

### Install

If you don't have Go installed yet, grab the latest stable version from https://golang.org/dl/ and install following instructions on https://golang.org/doc/install.   

Set GOPATH environment variable as described in [http://golang.org/doc/code.html#GOPATH](http://golang.org/doc/code.html#GOPATH) - for example `export GOPATH=$HOME/bin`.  (You may wish to add $GOPATH/bin to your PATH.) Then run:    
  
	$ go get github.com/yahoo/webseclab/...

### Run

```
$GOPATH/bin/webseclab [-http=:8080]
```
or simply ```webseclab``` if $GOPATH/bin is in your PATH.

Run webseclab -help to view the options.  

### Webseclab Tests

In all tests, excepts where specially mentioned, the attack input is assumed to be placed in the "in" CGI variable: &lt;url&gt;?in=&lt;attack_string&gt;. See the index page for PoEs (proof of exploits).

#### Reflected XSS

* xss/reflect/raw1 - echoes "raw" tags = literal '&lt;' and '&gt;' sent by the browser (IE-related). Can be tested with curl (Firefox/Chrome/Safari escape tag characters when sending to the server)

* xss/reflect/basic - echo of unfiltered input in a "normal" HTML context (not between tags, etc.). The example shows the minimal Webseclab template consisting of just {{.In}} placeholder.  PoE: /xss/reflect/basic?in=&lt;script&gt;alert(/HACKED/)&lt;/script&gt;  or /xss/reflect/basic?in=&lt;img src=foo onerror=alert(12345)&gt;

* xss/reflect/basic_in_tag - echo of unfiltered input inside of a "regular" HTML tag (&lt;B&gt;) PoE: /xss/reflect/basic_in_tag?in=&lt;script&gt;alert(/HACKED/)&lt;/script&gt;  or /xss/reflect/basic_in_tag?in=&lt;img src=foo onerror=alert(12345)&gt;

* xss/reflect/full1 - Javascript injection with closed quotes and a script tag echoed

* xss/reflect/post1 - same as above with injection via POST "in" form field (only POST method is allowed). xss/reflect/post1_splash can be used as a starting page with the action URL of xss/reflect/post1.

* xss/reflect/doubq1 - injection of double-escaped tags such as: xss/reflect/doubq1?in=%253Cscript%253Ealert%28%252FXSS%252F%29%253C%252Fscript%253E

* xss/reflect/rs1 - Response-Splitting attack, injection of %0D%0A%0D%0A which echoed unescaped in the header turning it into the response body. PoE:
/xss/reflect/rs1?in=xyz%0D%0A%0D%0A<script>alert(/BAD_NEWS/)</script>

* xss/reflect/onmouseover* - XSS due to attribute injections in tags (such as onmouseover handler)

* xss/reflect/oneclick1 - JS injection into JS executable context (unquoted input) - so-called "oneclick XSS".

* xss/reflect/refer -  the Referer header echoed. You can set up a page pointing to <WEBSECLAB_URL>/misc/webseclab_refer.html?%3Cscript%3Ealert%28789%29%3C/script%3E as a starting point to set the referer. 

* xss/reflect/js* - different cases of injection into Javascript blocks, see the index page for more details

* xss/reflect/enc2 - double quotes escaped with a backslash but backslash itself is not.  Exploitable injection into Javascript strings. 

* xss/reflect/backslash1?in=xyz - Unicode escape sequences like \u0022 unescaped by the server to became the corresponding (dangerous) character (double quotes). 

#### DOM XSS
* xss/dom/domwrite?in=foo - passing the unescaped document.location value to document.write(), PoE (Firefox): /xss/dom/domwrite?in=%3Cimg%20src=foo%20onerror=alert%28123%29%3E

* xss/dom/domwrite_hash?#whatever - passing the unescaped document.hash value to document.write(). PoE (Firefox): /xss/dom/domwrite_hash?#in=%3Cimg%20src=foo%20onerror=alert%281246%29%3E

* xss/dom/domwrite_hash_urlstyle#/foo/bar?in=whatever - passing the unescaped document.hash URL-style value to document.write(). PoE (Firefox): /xss/dom/domwrite_hash_urlstyle#/foo/bar?in=%3Cimg%20src=foo%20onerror=alert%281246%29%3E

* xss/dom/yuinode_hash?#in=xyz - passing the hash value to YUI's setHTML function.  PoE (Chrome/Firefox): /xss/dom/yuinode_hash?#in=xyz">/xss/dom/yuinode_hash?#in=xyz</A> - DOM XSS using YUI (location.hash) 

* xss/dom/yuinode_hash_urlstyle/#/foo/bar?in=xyz - passing the URL-style hash value to YUI's setHTML function.  PoE (Chrome/Firefox): /xss/dom/yuinode_hash_urlstyle/#/foo/bar?in=xyz">/xss/dom/yuinode_hash_urlstyle/#/foo/bar?in=xyz</A> - DOM XSS using YUI (location.hash, URL-style value) 

* xss/dom/yuinode_hash_unencoded?#in=xyz - passing the unencoded hash value to YUI's setHTML function.  PoE (Firefox / Chrome): /xss/dom/yuinode_hash?#in=xyz">/xss/dom/yuinode_hash?#in=xyz</A> - DOM XSS using YUI (decoded location.hash) 

### Adding New Tests

For most of the tests, you need to add a template that contains the "moustache" with {{.In}}.

To add a new test where input is echoed unfiltered, just drop an html
template under templates directory (for example templates/xss/newfile) with the template containing the {{.In}} placeholder.

To add a new "filter-based" case, add a template as above and add
a mapping of the corresponding entrypoint (such as /xss/newfile )
to the map in the FilterMap function in custom.go.  For example:  
```mp["/xss/reflect/newtest"] = []filter{TagsOff, SingleQuotesOff, GreaterThanOff}```  
 for a test with the corresponding input filtering.  See filters.go for the list of the available filters.

To add a new fully custom testcase, add a template (if needed),
add a mapping of the entrypoint to the handling function to CustomMap in custom.go and implement the custom function with the signature: func(http.ResponseWriter, *http.Request).  For example, for a test case with XSS injection through the Morse code, you could add:  
```mp["/xss/reflect/morse"] = XssUnsafeMorse```  

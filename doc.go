// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package webseclab contains a sample set of tests for web security scanners and a tooolkit to create such tests.

To install webseclab: go get github.com/yahoo/webseclab/...

To run the webseclab, execute the binary:

$ webseclab

By default, it will start the web server on port 8080, add -http=:<port> parameter to change it.

Run 'webseclab -help' for the usage help.

The convention is that the "in" cgi parameter (as in: ?in=foo)
is used to passed unsafe input (unless there are special functions
that pickup the input form other places like POST or headers - see custom.go)

See sample.html for an example of a template that can be used.

To add a new test where input is echoed unfiltered, just drop an html
template into somewhere under templates directory (for example templates/xss/newfile)
with the template containing the "moustache" with: {{.In}}

To add a new "filter-based" case, add a template as above and add
a mapping of the corresponding entrypoint (such as /xss/newfile )
to the map in the filterMap function.

To add a new fully custom testcase, add a template (if needed),
add the mapping of entrypoint to the handling function to CustomMap
and implement the custom function ( func(http.ResponseWriter, *http.Request) )
*/
package webseclab

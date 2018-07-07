// Copyright 2015, Yahoo Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// custom tests

package webseclab

import (
	"errors"
	"fmt"
	ht "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// CustomMap returns a map of entrypoint to handling functions
func CustomMap() (mp map[string]func(http.ResponseWriter, *http.Request) *LabResp) {
	mp = make(map[string]func(http.ResponseWriter, *http.Request) *LabResp)
	mp["/xss/reflect/backslash1"] = XSSBackslash
	mp["/xss/reflect/doubq1"] = XSSDoubq
	mp["/xss/reflect/enc2"] = XSSEnc
	mp["/xss/reflect/enc2_fp"] = XSSEncFp
	mp["/xss/reflect/full_cookies1"] = XSSFullCookies
	mp["/xss/reflect/full_headers1"] = XSSFullHeaders
	mp["/xss/reflect/full_useragent1"] = XSSFullUseragent
	mp["/xss/reflect/inredirect1_fp"] = XSSInRedirectFp
	mp["/xss/reflect/post1"] = XSSPost
	mp["/xss/reflect/refer1"] = XSSReferer
	mp["/xss/reflect/rs1"] = XSSRs

	return mp
}

// filterMap returns a map of entrypoint to a slice of filtering options
func filterMap() (mp map[string][]filter) {
	mp = make(map[string][]filter)
	mp["/misc/escapeexample_nogt"] = []filter{GreaterThanOff}
	mp["/misc/escapeexample_nogt_noquotes"] = []filter{QuotesOff, GreaterThanOff}
	mp["/xss/reflect/enc2"] = []filter{TagsOff, DoubleQuotesBackslashEscape}
	mp["/xss/reflect/enc2_fp"] = []filter{TagsOff, DoubleQuotesBackslashEscape, BackslashEscape}
	mp["/xss/reflect/js3"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/js3_fp"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/js3_notags"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/js3_notags_fp"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/js3_search_fp"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/js4_dq"] = []filter{TagsOff, SingleQuotesOff}
	mp["/xss/reflect/js4_dq_fp"] = []filter{TagsOff, BackslashEscapeDoubleQuotesAndBackslash}
	mp["/xss/reflect/js6_bug7208690"] = []filter{TagsOff, DoubleQuotesOff}
	mp["/xss/reflect/js6_sq"] = []filter{TagsOff, DoubleQuotesOff}
	mp["/xss/reflect/js6_sq_combo1"] = []filter{TagsOff, DoubleQuotesOff}
	mp["/xss/reflect/js6_sq_fp"] = []filter{TagsOff, DoubleQuotesOff}
	mp["/xss/reflect/js_script_close"] = []filter{QuotesOff}
	mp["/xss/reflect/oneclick1"] = []filter{QuotesOff, TagsOff}
	mp["/xss/reflect/onmouseover"] = []filter{TagsOff}
	mp["/xss/reflect/onmouseover_div_unquoted"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/onmouseover_div_unquoted_fp"] = []filter{TagsOff, QuotesOff, SpacesOff}
	mp["/xss/reflect/onmouseover_unquoted"] = []filter{TagsOff, QuotesOff}
	mp["/xss/reflect/onmouseover_unquoted_fp"] = []filter{TagsOff, QuotesOff, SpacesOff}
	mp["/xss/reflect/raw1_fp"] = []filter{QuotesOff, ScriptOff}
	mp["/xss/reflect/textarea1"] = []filter{TagsOffUntilTextareaClose}
	mp["/xss/reflect/textarea1_fp"] = []filter{TextareaCloseOff}
	mp["/xss/reflect/textarea2"] = []filter{NoOp}
	mp["/xss/reflect/textarea2_fp"] = []filter{TextareaSafe}
	return
}

// XSSRs filters output to produce request splitting (injection of HEAD/BODY separator).
func XSSRs(w http.ResponseWriter, r *http.Request) *LabResp {
	r.Header.Write(os.Stdout)
	in := r.FormValue("in")
	if in == "" {
		w.Header().Set("Content-type", "text/html; charset=utf-8")
		w.Write([]byte(`No &quot;in&quot; cgi parameter passed - nothing to inject!`))
		return &LabResp{Err: nil, Code: http.StatusOK}
	}

	safe := strings.Replace(in, "\x0A", "%0a", -1)
	safe = strings.Replace(safe, "\x0D", "%0d", -1)

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return &LabResp{Err: errors.New("Bad Error in Hijacking (00)"), Code: http.StatusInternalServerError}
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return &LabResp{Err: errors.New("Bad Error in Hijacking (01)"), Code: http.StatusInternalServerError}
	}
	// Don't forget to close the connection:
	defer conn.Close()
	bufrw.WriteString("HTTP/1.1 200\n")
	bufrw.WriteString("Content-type: text/html; charset=utf-8\n")
	bufrw.WriteString("X-Test: filtered value in the HTTP header (CR and LF removed): " + safe + "\n")
	if !strings.HasSuffix(r.URL.Path, ".ok") {
		bufrw.WriteString("X-Foo-Unsafe: unfiltered value in the HTTP header (CR and LF not removed): " + in + "\n")
	}
	bufrw.WriteString("X-Hopefully-Header: with the correct input handling , this should still be in the HTTP header\n")
	bufrw.WriteByte(10)
	bufrw.WriteByte(13)
	bufrw.WriteByte(10)
	bufrw.WriteByte(13)
	tmplsrc := `<!doctype html><H1>Webseclab</H1><p><H2>Request splitting test</H2><p>The input parameters in the request body should be propertly filtered: <BR>in = {{.In}}</html>`
	tmpl, err := ht.New("test").Parse(string(tmplsrc))
	input := InData{In: in}
	err = tmpl.Execute(bufrw, input)
	if err != nil {
		panic(err)
	}
	bufrw.WriteString("</body></html>\n")
	bufrw.Flush()
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSFullCookies is a wrapper around standard handler (non-filtered output echo)
// but requires the presence of a cookie with value "awesome"
func XSSFullCookies(w http.ResponseWriter, r *http.Request) *LabResp {
	for _, ck := range r.Cookies() {
		if strings.Contains(ck.Value, "awesome") {
			return DoLabTestStandard(w, r)
		}
	}
	return &LabResp{Err: errors.New("sorry your cookies are no good - please put the word awesome into a cookie value and try again"),
		Code: http.StatusForbidden}
}

// XSSFullHeaders is a wrapper around standard handler (non-filtered output echo)
// but requires the presence of HTTP Header X-Letmein with the value 1.
func XSSFullHeaders(w http.ResponseWriter, r *http.Request) *LabResp {
	if r.Header.Get("X-Letmein") != "1" {
		return &LabResp{Err: errors.New("missing or invalid value of the X-Letmein HTTP Header - please set to 1 and try again"),
			Code: http.StatusForbidden}
	}
	return DoLabTestStandard(w, r)
}

// XSSFullUseragent is a wrapper around standard handler (non-filtered output echo)
// but requires the presence of a Header "User-Agent" with substring "Mobile" in the value
func XSSFullUseragent(w http.ResponseWriter, r *http.Request) *LabResp {
	ua := r.Header.Get("User-Agent")
	if !strings.Contains(ua, "Mobile") {
		return &LabResp{Err: errors.New("access requires forward-looking thinking - please add Mobile to the User-Agent header and try again"),
			Code: http.StatusForbidden}
	}
	return DoLabTestStandard(w, r)
}

// XSSPost handles POST input.
func XSSPost(w http.ResponseWriter, r *http.Request) *LabResp {
	var inp InData
	if r.Method == "GET" {
		err := DoTemplate(w, r.URL.Path, &InData{In: "wrong Method - expecting POST"})
		if err != nil {
			log.Printf("Error in XssPost: %s\n", err)
			return &LabResp{Err: err, Code: http.StatusInternalServerError}
		}
		return &LabResp{Err: nil, Code: http.StatusOK}
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body in XssPost01: %s\n", err)
		return &LabResp{Err: err, Code: http.StatusInternalServerError}
	}
	bodyUnescaped, err := url.QueryUnescape(string(body))
	if err != nil {
		log.Printf("Error unescaping request body in XssPost01: %s\n", err)
		return &LabResp{Err: err, Code: http.StatusInternalServerError}
	}
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, bodyUnescaped)

	if in, ok := rawParams["in"]; ok {
		inp.In = in[0]
	}

	err = DoTemplate(w, r.URL.Path, &inp)
	if err != nil {
		log.Printf("Error in XssPost1: %s\n", err)
		return &LabResp{Err: err, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSReferer copies and echoes the Referer header.
func XSSReferer(w http.ResponseWriter, r *http.Request) *LabResp {
	inp := &InData{} // placeholder
	referer := r.Header.Get("Referer")
	inp.InRaw = referer
	refererUnescaped, err := url.QueryUnescape(referer)
	if err != nil {
		log.Printf("Error in XssReferer - unable to QueryUnescape %s - %s\n", referer, err)
		inp.In = referer
	} else {
		inp.In = refererUnescaped
	}
	err = DoTemplate(w, r.URL.Path, inp)
	if err != nil {
		log.Printf("Error in XssReferer: %s\n", err)
		return &LabResp{Err: err, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSInRedirectFp issues a redirect to Yahoo homepage.
func XSSInRedirectFp(w http.ResponseWriter, r *http.Request) *LabResp {
	w.Header().Set("Location", "https://www.yahoo.com")
	w.WriteHeader(http.StatusFound)
	input := &InData{}
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
		unesc, err := url.QueryUnescape(input.InRaw)
		if err != nil {
			fmt.Printf("ERROR in url.QueryUnescape on %s\n", input.InRaw)
		}
		input.In = unesc
	}
	err := DoTemplate(w, r.URL.Path, input)
	if err != nil {
		log.Printf("Error in DoTemplate: %s\n", err)
		return &LabResp{Err: nil, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSEnc escapes quotes with backslash but does not escape backslash itself
// allowing injection of an unescaped double quote
func XSSEnc(w http.ResponseWriter, r *http.Request) *LabResp {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	input := &InData{}
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
		unesc, err := url.QueryUnescape(input.InRaw)
		if err != nil {
			log.Printf("Error in XssEnc2 / QueryUnescape: %s\n", err)
			return &LabResp{Err: nil, Code: http.StatusInternalServerError}
		}
		input.In = Transform(unesc, TagsOff, DoubleQuotesBackslashEscape)
	}
	err := DoTemplate(w, r.URL.Path, input)
	if err != nil {
		log.Printf("Error in DoTemplate: %s\n", err)
		return &LabResp{Err: nil, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSEncFp escapes quotes and backslash with backslash preventing injection
func XSSEncFp(w http.ResponseWriter, r *http.Request) *LabResp {
	input := &InData{}
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
		unesc, err := url.QueryUnescape(input.InRaw)
		if err != nil {
			log.Printf("Error in XssEnc2 / QueryUnescape: %s\n", err)
			return &LabResp{Err: nil, Code: http.StatusInternalServerError}
		}
		input.In = Transform(unesc, TagsOff, QuotesOff)
	}
	err := DoTemplate(w, r.URL.Path, input)
	if err != nil {
		log.Printf("Error in DoTemplate: %s\n", err)
		return &LabResp{Err: nil, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSDoubq double-unencodes the "in" cgi parameter.
func XSSDoubq(w http.ResponseWriter, r *http.Request) *LabResp {
	input := &InData{}
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		// imitate a bad way to do input validation - single level unescape
		input.InRaw = inputRaw[0]
		// one-level escape
		unesc1, err := url.QueryUnescape(inputRaw[0])
		if err != nil {
			fmt.Printf("ERROR in the first url.QueryUnescape on %s\n", inputRaw[0])
			return &LabResp{Err: nil, Code: http.StatusBadRequest}
		}
		unesc1 = Transform(unesc1, TagsOff, QuotesOff)
		unesc, err := url.QueryUnescape(unesc1)
		if err != nil {
			fmt.Printf("ERROR in the second url.QueryUnescape on %s\n", unesc1)
			return &LabResp{Err: nil, Code: http.StatusBadRequest}
		}
		input.In = unesc
	}
	err := DoTemplate(w, r.URL.Path, input)
	if err != nil {
		log.Printf("Error in DoTemplate: %s\n", err)
		return &LabResp{Err: nil, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

// XSSBackslash is a special function that uses UnescapeUnicode to convert
// \u{dd} into the {dd} ASCII character.
func XSSBackslash(w http.ResponseWriter, r *http.Request) *LabResp {
	input := &InData{}
	rawParams := make(map[string][]string)
	ParseRawQuery(rawParams, r.URL.RawQuery)
	inputRaw, ok := rawParams["in"]
	if ok && len(inputRaw) > 0 {
		input.InRaw = inputRaw[0]
		input.In = Transform(input.InRaw, TagsOff, QuotesOff)
		input.In = UnescapeUnicode(input.In)
	}
	err := DoTemplate(w, r.URL.Path, input)
	if err != nil {
		log.Printf("Error in DoTemplate: %s\n", err)
		return &LabResp{Err: nil, Code: http.StatusInternalServerError}
	}
	return &LabResp{Err: nil, Code: http.StatusOK}
}

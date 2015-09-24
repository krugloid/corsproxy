package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func r(method string, url string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	return req
}

func testServer(req *http.Request, fn http.HandlerFunc) *httptest.ResponseRecorder {
	ts := httptest.NewServer(fn)
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	req.URL.Scheme = "http"
	req.URL.Path = "/http://" + u.Host + req.URL.Path
	req.URL.Host = "example.com"
	w := httptest.NewRecorder()
	(&corsHandler{}).ServeHTTP(w, req)
	return w
}

// Test if response contains CORS headers
func TestCORS(t *testing.T) {
	w := testServer(r("GET", "/foo", nil), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	h := w.Header()
	if _, ok := h["Access-Control-Allow-Origin"]; !ok {
		t.Error(h)
	}
	if _, ok := h["Access-Control-Allow-Headers"]; !ok {
		t.Error(h)
	}
	if _, ok := h["Access-Control-Allow-Credentials"]; !ok {
		t.Error(h)
	}
	if _, ok := h["Access-Control-Allow-Methods"]; !ok {
		t.Error(h)
	}
}

// Test if request URL path and query are passed to the target server
func TestRequestURL(t *testing.T) {
	req := r("GET", "/foo", nil)
	v := req.URL.Query()
	v.Add("a", "b")
	v.Add("c", "123")
	req.URL.RawQuery = v.Encode()
	testServer(req, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/foo" {
			t.Error(r.URL)
		}
		q := r.URL.Query()
		if len(q) != 2 || q.Get("a") != "b" || q.Get("c") != "123" {
			t.Error(q)
		}
	})
}

// Test if request headers are passed to the target server
func TestRequestHeaders(t *testing.T) {
	req := r("GET", "/foo", nil)
	req.Header.Add("X-TestHeader", "foo")
	testServer(req, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-TestHeader") != "foo" {
			t.Error(r.Header)
		}
	})
}

// Test if request body is passed to the target server
func TestRequestBody(t *testing.T) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, _ := ioutil.ReadAll(r.Body)
		if string(b) != "bar" {
			t.Error(b)
		}
	}
	testServer(r("POST", "/foo", bytes.NewBufferString("bar")), fn)
}

// Test if target server response status is passed back
func TestResponseStatusCode(t *testing.T) {
	w := testServer(r("GET", "/foo", nil), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	if w.Code != 404 {
		t.Error(w)
	}
}

// Test if target server response headers are passed back
func TestResponseHeaders(t *testing.T) {
	w := testServer(r("GET", "/foo", nil), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-TestHeader", "bar")
		w.WriteHeader(200)
	})
	if w.Header().Get("X-TestHeader") != "bar" {
		t.Error(w)
	}
}

// Test if target server response body is passed back
func TestResponseBody(t *testing.T) {
	w := testServer(r("GET", "/foo", nil), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})
	if w.Body.String() != "Hello world" {
		t.Error(w)
	}
}

// Test malformed URL
func TestBadURL(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/foo://bar", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	(&corsHandler{}).ServeHTTP(w, req)
	if w.Code != 400 {
		t.Error(w)
	}
}

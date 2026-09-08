package tls_client_cffi_src

import (
	"testing"

	http "github.com/bogdanfinn/fhttp"
)

// The client's default headers stand in for an empty header set, so a request
// built from an input with no headers has to have an empty header map. It used
// to carry the header order key with nothing in it, and that one entry was
// enough to keep defaultHeaders from ever applying through the shared library.
func TestBuildRequestWithoutHeadersLeavesTheMapEmpty(t *testing.T) {
	req, cerr := BuildRequest(RequestInput{RequestMethod: "GET", RequestUrl: "https://example.test/"})
	if cerr != nil {
		t.Fatal(cerr)
	}
	if len(req.Header) != 0 {
		t.Fatalf("an input without headers gives a request with %d header entries: %v", len(req.Header), req.Header)
	}

	// The control: an order that was given is kept.
	req, cerr = BuildRequest(RequestInput{
		RequestMethod: "GET", RequestUrl: "https://example.test/",
		Headers: map[string]string{"a": "1", "b": "2"}, HeaderOrder: []string{"b", "a"},
	})
	if cerr != nil {
		t.Fatal(cerr)
	}
	if got := req.Header[http.HeaderOrderKey]; len(got) != 2 || got[0] != "b" {
		t.Errorf("the header order was not kept: %v", got)
	}
}

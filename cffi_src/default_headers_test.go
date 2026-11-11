package tls_client_cffi_src

import (
	"strings"
	"testing"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
)

// A request that names a browser profile and sets no headers goes out with the
// browser's headers; a request with a header of its own, or with a profile that
// is not a mapped browser, is left as it was.
func TestABrowserProfileWithNoHeadersSendsTheBrowsersHeaders(t *testing.T) {
	// The handler runs on the server's goroutine, so what it saw travels
	// over a channel instead of a shared variable.
	got := make(chan http.Header, 1)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.Header.Clone()
	}))
	defer srv.Close()

	send := func(t *testing.T, input RequestInput) http.Header {
		t.Helper()
		input.RequestMethod = "GET"
		input.RequestUrl = srv.URL
		input.InsecureSkipVerify = true
		input.WithoutCookieJar = true
		input.TimeoutSeconds = 10
		client, _, _, cerr := CreateClient(input)
		if cerr != nil {
			t.Fatal(cerr)
		}
		req, cerr := BuildRequest(input)
		if cerr != nil {
			t.Fatal(cerr)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		select {
		case h := <-got:
			return h
		case <-time.After(5 * time.Second):
			t.Fatal("the server saw no request")
			return nil
		}
	}

	h := send(t, RequestInput{TLSClientIdentifier: "chrome_152"})
	if ua := h.Get("User-Agent"); !strings.Contains(ua, "Chrome/152.0.0.0") {
		t.Errorf("chrome_152 with no headers sent User-Agent %q", ua)
	}
	if h.Get("sec-ch-ua") == "" || h.Get("Accept-Language") == "" {
		t.Errorf("chrome_152 with no headers sent an incomplete set: %v", h)
	}

	h = send(t, RequestInput{TLSClientIdentifier: "chrome_152", Headers: map[string]string{"user-agent": "mine"}})
	if ua := h.Get("User-Agent"); ua != "mine" {
		t.Errorf("a request with its own User-Agent sent %q", ua)
	}
	if h.Get("sec-ch-ua") != "" {
		t.Errorf("a request with its own headers had the browser's added: %v", h)
	}

	h = send(t, RequestInput{TLSClientIdentifier: "firefox_133"})
	if ua := h.Get("User-Agent"); strings.Contains(ua, "Chrome") {
		t.Errorf("firefox_133 has no mapped headers and still sent %q", ua)
	}
}

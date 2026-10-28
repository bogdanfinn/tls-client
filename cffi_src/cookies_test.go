package tls_client_cffi_src

import (
	"net/url"
	"testing"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	tls_client "github.com/bogdanfinn/tls-client"
)

func newSessionWithJar(t *testing.T) tls_client.HttpClient {
	t.Helper()
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), tls_client.WithCookieJar(tls_client.NewCookieJar()))
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// as a map of name to value per host, so the comparison ignores the order the
// jar happens to store cookies in.
func flatten(all map[string][]Cookie) map[string]map[string]string {
	out := make(map[string]map[string]string)
	for host, cookies := range all {
		out[host] = make(map[string]string)
		for _, c := range cookies {
			out[host][c.Name] = c.Value
		}
	}
	return out
}

func TestAllCookiesFromSessionExportsTheWholeJar(t *testing.T) {
	client := newSessionWithJar(t)
	a, _ := url.Parse("https://example.com/")
	b, _ := url.Parse("https://other.org/")
	client.SetCookies(a, []*http.Cookie{{Name: "s", Value: "1"}, {Name: "t", Value: "2"}})
	client.SetCookies(b, []*http.Cookie{{Name: "u", Value: "3"}})

	all, err := AllCookiesFromSession(client)
	if err != nil {
		t.Fatal(err)
	}

	got := flatten(all)
	if len(got) != 2 {
		t.Fatalf("exported %d hosts, want 2: %v", len(got), got)
	}
	if got["example.com"]["s"] != "1" || got["example.com"]["t"] != "2" {
		t.Errorf("example.com: %v", got["example.com"])
	}
	if got["other.org"]["u"] != "3" {
		t.Errorf("other.org: %v", got["other.org"])
	}

	// getCookiesFromSession only sees one URL; the control is that this saw both.
	if only := client.GetCookies(a); len(only) != 2 {
		t.Errorf("the single-URL read returned %d cookies, so the two-host export is the new part", len(only))
	}
}

func TestSetCookiesToSessionRoundTrips(t *testing.T) {
	src := newSessionWithJar(t)
	a, _ := url.Parse("https://example.com/")
	b, _ := url.Parse("https://other.org/")
	src.SetCookies(a, []*http.Cookie{{Name: "s", Value: "1"}, {Name: "t", Value: "2"}})
	src.SetCookies(b, []*http.Cookie{{Name: "u", Value: "3"}})

	exported, err := AllCookiesFromSession(src)
	if err != nil {
		t.Fatal(err)
	}

	dst := newSessionWithJar(t)
	after, err := SetCookiesToSession(dst, exported)
	if err != nil {
		t.Fatal(err)
	}

	// The returned state and a fresh read of the destination agree, and both
	// equal what was exported.
	reread, err := AllCookiesFromSession(dst)
	if err != nil {
		t.Fatal(err)
	}
	want := flatten(exported)
	for label, got := range map[string]map[string]map[string]string{"returned": flatten(after), "reread": flatten(reread)} {
		if len(got) != len(want) {
			t.Fatalf("%s: %d hosts, want %d", label, len(got), len(want))
		}
		for host, cookies := range want {
			for name, value := range cookies {
				if got[host][name] != value {
					t.Errorf("%s: %s %s = %q, want %q", label, host, name, got[host][name], value)
				}
			}
		}
	}
}

func TestAllCookiesRefusesAJarItCannotEnumerate(t *testing.T) {
	// The control is the tls-client jar above, which enumerates. A bare fhttp
	// jar cannot, and both functions say so instead of returning nothing.
	plain, _ := cookiejar.New(nil)
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), tls_client.WithCookieJar(plain))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := AllCookiesFromSession(client); err == nil {
		t.Error("AllCookiesFromSession did not refuse a jar it cannot enumerate")
	}
	if _, err := SetCookiesToSession(client, map[string][]Cookie{"example.com": {{Name: "s", Value: "1"}}}); err == nil {
		t.Error("SetCookiesToSession did not refuse a jar it cannot enumerate")
	}
}

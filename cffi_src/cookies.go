package tls_client_cffi_src

import (
	"fmt"
	"net/url"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

// AllCookiesFromSession returns every cookie the client's jar holds, grouped by
// the jar's host key, so a whole session can be exported and later restored with
// SetCookiesToSession. getCookiesFromSession only returns the cookies that would
// be sent to one URL; this returns the entire jar.
//
// It needs the tls-client cookie jar, which is what NewCookieJar and the shared
// library's withCustomCookieJar build. A bare fhttp jar keeps no enumerable copy
// of its cookies and is refused with a clear error rather than silently returning
// nothing.
func AllCookiesFromSession(client tls_client.HttpClient) (map[string][]Cookie, error) {
	jar, ok := client.GetCookieJar().(tls_client.CookieJar)
	if !ok {
		return nil, fmt.Errorf("the session's cookie jar cannot enumerate its cookies; create the session with a tls-client cookie jar (withCustomCookieJar) to export all cookies")
	}

	out := make(map[string][]Cookie)
	for host, cookies := range jar.GetAllCookies() {
		out[host] = cookiesToCffi(cookies)
	}

	return out, nil
}

// SetCookiesToSession imports cookies grouped by host, the shape
// AllCookiesFromSession returns, into the client's jar and returns the jar's
// contents afterwards. Each host group is set under a URL built from that host,
// so a jar exported from one session and imported into another holds the same
// cookies under the same keys.
//
// It has the same jar requirement as AllCookiesFromSession, checked before any
// cookie is set, so a session that cannot enumerate is refused rather than
// half-importing.
func SetCookiesToSession(client tls_client.HttpClient, cookies map[string][]Cookie) (map[string][]Cookie, error) {
	if _, ok := client.GetCookieJar().(tls_client.CookieJar); !ok {
		return nil, fmt.Errorf("the session's cookie jar cannot enumerate its cookies; create the session with a tls-client cookie jar (withCustomCookieJar) to import all cookies")
	}

	for host, hostCookies := range cookies {
		u := &url.URL{Scheme: "https", Host: host, Path: "/"}
		client.SetCookies(u, cookiesFromCffi(hostCookies))
	}

	return AllCookiesFromSession(client)
}

func cookiesToCffi(cookies []*http.Cookie) []Cookie {
	out := make([]Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		out = append(out, Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			MaxAge:   cookie.MaxAge,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
			Expires:  Timestamp{Time: cookie.Expires},
		})
	}

	return out
}

func cookiesFromCffi(cookies []Cookie) []*http.Cookie {
	out := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		out = append(out, &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires.Time,
			MaxAge:   cookie.MaxAge,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
		})
	}

	return out
}

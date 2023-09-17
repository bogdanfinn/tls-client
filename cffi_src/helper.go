package tls_client_cffi_src

import (
	"fmt"
	http "github.com/bogdanfinn/fhttp"
)

func BuildCookies(cookies []Cookie) []*http.Cookie {
	var ret []*http.Cookie

	for _, cookie := range cookies {
		ret = append(ret, &http.Cookie{
			Name:    cookie.Name,
			Value:   cookie.Value,
			Path:    cookie.Path,
			Domain:  cookie.Domain,
			Expires: cookie.Expires.Time,
		})
	}

	return ret
}

func ToCookieMap(cookies []*http.Cookie) map[string][]*http.Cookie {
	ret := make(map[string][]*http.Cookie)

	for _, cookie := range cookies {
		urlString := fmt.Sprintf("%s%s", cookie.Domain, cookie.Path)

		ret[urlString] = append(ret[urlString], cookie)
	}

	return ret
}

func ToCookieSlice(cookies map[string][]*http.Cookie) []*http.Cookie {
	var ret []*http.Cookie

	for _, cookies := range cookies {
		ret = append(ret, cookies...)
	}

	return ret
}

func TransformCookies(cookies []*http.Cookie) []Cookie {
	var ret []Cookie

	for _, cookie := range cookies {
		ret = append(ret, Cookie{
			Name:   cookie.Name,
			Value:  cookie.Value,
			Path:   cookie.Path,
			Domain: cookie.Domain,
			Expires: Timestamp{
				Time: cookie.Expires,
			},
		})
	}

	return ret
}

package tls_client

import (
	"strings"
	"sync"
)

type BetterJar struct {
	cookies map[string]string
	mu      sync.RWMutex
}

func NewBetterJar() *BetterJar {
	return &BetterJar{
		cookies: make(map[string]string),
	}
}

func (bj *BetterJar) SetCookies(cookieString string) {
	bj.mu.Lock()
	defer bj.mu.Unlock()

	cookies := strings.Split(cookieString, ";")
	for _, cookie := range cookies {
		nameI := strings.Index(cookie, "=")
		if nameI == -1 {
			continue
		}
		name := strings.TrimSpace(cookie[:nameI])
		value := strings.TrimSpace(cookie[nameI+1:])

		if shouldProcessCookie(name, value) {
			bj.cookies[name] = value
		}
	}
}

func (bj *BetterJar) GetCookies() string {
	bj.mu.RLock()
	defer bj.mu.RUnlock()

	cookies := ""
	for name, value := range bj.cookies {
		if shouldProcessCookie(name, value) {
			cookies += name + "=" + value + ";"
		}
	}

	return strings.TrimSuffix(cookies, ";")
}
func (bj *BetterJar) processCookies(resp *WebResp) {
	setCookies := resp.Header.Values("Set-Cookie")

	if len(setCookies) == 0 {
		resp.Cookies = bj.GetCookies()
		return
	}

	bj.mu.Lock()
	defer bj.mu.Unlock()

	for _, setCookie := range setCookies {
		cookieAttributes := strings.Split(setCookie, ";")

		// Parse and process each attribute
		var found = false
		for _, attr := range cookieAttributes {
			if found {
				break
			}
			attr = strings.TrimSpace(attr)
			parts := strings.SplitN(attr, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			switch strings.ToLower(name) {
			case "path", "domain", "expires":
				continue
			default:
				if shouldProcessCookie(name, value) {
					bj.cookies[name] = value
				}
				found = true
			}
		}
	}

	resp.Cookies = bj.GetCookies()
}

func shouldProcessCookie(name, value string) bool {
	return name != "" && value != "" && value != `""` && value != "undefined"
}

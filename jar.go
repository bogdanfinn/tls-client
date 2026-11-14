package tls_client

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
)

type CookieJarOption func(config *cookieJarConfig)

type cookieJarConfig struct {
	logger            Logger
	skipExisting      bool
	debug             bool
	allowEmptyCookies bool
}

func WithSkipExisting() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.skipExisting = true
	}
}

func WithAllowEmptyCookies() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.allowEmptyCookies = true
	}
}

func WithDebugLogger() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.debug = true
	}
}

func WithLogger(logger Logger) CookieJarOption {
	return func(config *cookieJarConfig) {
		config.logger = logger
	}
}

type CookieJar interface {
	http.CookieJar
	GetAllCookies() map[string][]*http.Cookie
}

type cookieJar struct {
	jar        *cookiejar.Jar
	config     *cookieJarConfig
	allCookies map[string][]*http.Cookie
	sync.RWMutex
}

func NewCookieJar(options ...CookieJarOption) CookieJar {
	realJar, _ := cookiejar.New(nil)

	config := &cookieJarConfig{}

	for _, opt := range options {
		opt(config)
	}

	if config.logger == nil {
		config.logger = NewNoopLogger()
	}

	if config.debug {
		config.logger = NewDebugLogger(config.logger)
	}

	c := &cookieJar{
		jar:        realJar,
		config:     config,
		allCookies: make(map[string][]*http.Cookie),
	}

	return c
}

func (jar *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Lock()
	defer jar.Unlock()

	notEmptyCookies := jar.nonEmpty(cookies)
	uniqueCookies := jar.unique(notEmptyCookies)

	// The standard jar applies the RFC 6265 domain rules and drops a cookie
	// whose Domain does not match the host it arrives from. allCookies keeps
	// the same set so GetAllCookies never reports a cookie the request path
	// could not send.
	storableCookies := jar.domainMatches(u, uniqueCookies)

	hostKey := jar.buildCookieHostKey(u)

	if !jar.config.skipExisting {
		existingCookies := jar.allCookies[hostKey]

		var remainingExistingCookies []*http.Cookie

		for _, existingCookie := range existingCookies {
			shouldOverwrite := false
			for _, cookie := range storableCookies {
				shouldOverwrite = existingCookie.Name == cookie.Name

				if shouldOverwrite {
					break
				}
			}

			if shouldOverwrite {
				continue
			}

			remainingExistingCookies = append(remainingExistingCookies, existingCookie)
		}

		newCookies := append(remainingExistingCookies, storableCookies...)

		// Only the newly arrived cookies go into the standard jar, each under
		// the URL it actually arrived with. Re-registering existing cookies
		// under a different URL of the same host key would widen a host-only
		// cookie to the parent domain.
		jar.jar.SetCookies(u, storableCookies)
		jar.allCookies[hostKey] = newCookies

		return
	}

	var newNonExistentCookies []*http.Cookie

	existingCookies := jar.allCookies[hostKey]

	for _, cookie := range storableCookies {
		alreadyInJar := false

		for _, existingCookie := range existingCookies {
			alreadyInJar = cookie.Name == existingCookie.Name

			if alreadyInJar {
				break
			}
		}

		if alreadyInJar {
			jar.config.logger.Debug("cookie %s is already in jar, skipping", cookie.Name)
			continue
		}

		jar.config.logger.Debug("cookie %s is not in jar yet, adding", cookie.Name)
		newNonExistentCookies = append(newNonExistentCookies, cookie)
	}

	jar.jar.SetCookies(u, newNonExistentCookies)
	jar.allCookies[hostKey] = append(existingCookies, newNonExistentCookies...)
}

func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	// The standard jar is the authority on which cookies may go back to u: it
	// applies the RFC 6265 domain, path, secure and expiry rules. allCookies
	// only tracks what was offered, so without this filter a cookie set for a
	// different host would leak into this request.
	standard := jar.jar.Cookies(u)

	sendable := make(map[string]struct{}, len(standard))
	for _, c := range standard {
		sendable[c.Name+"\x00"+c.Value] = struct{}{}
	}

	var filteredCookies []*http.Cookie

	for _, c := range jar.notExpired(jar.allCookies[jar.buildCookieHostKey(u)]) {
		key := c.Name + "\x00" + c.Value
		if _, ok := sendable[key]; ok {
			delete(sendable, key)
			filteredCookies = append(filteredCookies, c)
		}
	}

	return filteredCookies
}

// domainMatches drops the cookies the standard jar would reject on arrival:
// a Domain attribute that is malformed or does not domain-match the host the
// cookie is set from (RFC 6265 section 5.3). A cookie without a Domain
// attribute is a host cookie and always allowed.
func (jar *cookieJar) domainMatches(u *url.URL, cookies []*http.Cookie) []*http.Cookie {
	host := strings.ToLower(u.Hostname())
	isIPHost := net.ParseIP(host) != nil

	var filteredCookies []*http.Cookie

	for _, c := range cookies {
		if c.Domain == "" {
			filteredCookies = append(filteredCookies, c)
			continue
		}

		if isIPHost {
			jar.config.logger.Debug("cookie %s carries a domain attribute on an IP host, dropping", c.Name)
			continue
		}

		domain := strings.ToLower(strings.TrimPrefix(c.Domain, "."))
		if domain == "" || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
			jar.config.logger.Debug("cookie %s carries a malformed domain %q, dropping", c.Name, c.Domain)
			continue
		}

		if host != domain && !strings.HasSuffix(host, "."+domain) {
			jar.config.logger.Debug("cookie %s carries domain %s which does not match host %s, dropping", c.Name, c.Domain, host)
			continue
		}

		filteredCookies = append(filteredCookies, c)
	}

	return filteredCookies
}

func (jar *cookieJar) GetAllCookies() map[string][]*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	copied := make(map[string][]*http.Cookie)
	for u, c := range jar.allCookies {
		copied[u] = c
	}

	return copied
}

func (jar *cookieJar) buildCookieHostKey(u *url.URL) string {
	host := u.Host

	hostParts := strings.Split(host, ".")

	switch len(hostParts) {
	case 3:
		return fmt.Sprintf("%s.%s", hostParts[len(hostParts)-2], hostParts[len(hostParts)-1])
	case 2:
		return fmt.Sprintf("%s.%s", hostParts[len(hostParts)-2], hostParts[len(hostParts)-1])
	default:
		return host
	}
}

func (jar *cookieJar) unique(cookies []*http.Cookie) []*http.Cookie {
	var filteredCookies []*http.Cookie
	var uniqueCookies []string

	for i := len(cookies) - 1; i >= 0; i-- {
		c := cookies[i]

		if inSlice(uniqueCookies, c.Name) {
			continue
		}

		filteredCookies = append(filteredCookies, c)
		uniqueCookies = append(uniqueCookies, c.Name)
	}

	return filteredCookies
}

func (jar *cookieJar) nonEmpty(cookies []*http.Cookie) []*http.Cookie {
	if jar.config.allowEmptyCookies {
		return cookies
	}

	var filteredCookies []*http.Cookie

	for _, c := range cookies {
		if c.Value == "" {
			jar.config.logger.Debug("cookie %s is empty and will be filtered out", c.Name)
			continue
		}

		filteredCookies = append(filteredCookies, c)
	}

	return filteredCookies
}

func (jar *cookieJar) notExpired(cookies []*http.Cookie) []*http.Cookie {
	var filteredCookies []*http.Cookie

	for _, c := range cookies {
		// we misuse the max age here for "deletion" reasons. To be 100% correct a MaxAge equals 0 should also be deleted but we do not do it for now.
		if c.MaxAge < 0 {
			jar.config.logger.Debug("cookie %s in jar max age set to 0 or below. will be excluded from request", c.Name)
			continue
		}

		// TODO: this is currently commented out as the cookie parser does not parse the expire correctly out of the Set-Cookie header.
		/*if c.Expires.Before(now) {
			jar.config.logger.Debug("cookie %s in jar expired. will be excluded from request", c.Name)
			continue
		}*/

		filteredCookies = append(filteredCookies, c)
	}

	return filteredCookies
}

func inSlice(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}

	return false
}

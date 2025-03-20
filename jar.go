package tls_client

import (
	"fmt"
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

	hostKey := jar.buildCookieHostKey(u)

	if !jar.config.skipExisting {
		existingCookies := jar.allCookies[hostKey]

		var remainingExistingCookies []*http.Cookie

		for _, existingCookie := range existingCookies {
			shouldOverwrite := false
			for _, cookie := range uniqueCookies {
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

		newCookies := append(remainingExistingCookies, uniqueCookies...)

		jar.jar.SetCookies(u, newCookies)
		jar.allCookies[hostKey] = newCookies

		return
	}

	var newNonExistentCookies []*http.Cookie

	existingCookies := jar.allCookies[hostKey]

	for _, cookie := range uniqueCookies {
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

	newCookies := append(existingCookies, newNonExistentCookies...)
	jar.jar.SetCookies(u, newCookies)
	jar.allCookies[hostKey] = newCookies
}

func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	hostKey := jar.buildCookieHostKey(u)

	allCookies := jar.allCookies[hostKey]

	return jar.notExpired(allCookies)
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

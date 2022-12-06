package tls_client

import (
	"net/url"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
)

type CookieJarOption func(config *cookieJarConfig)

type cookieJarConfig struct {
	skipExisting bool
	logger       Logger
}

func WithSkipExisting() CookieJarOption {
	return func(config *cookieJarConfig) {
		config.skipExisting = true
	}
}

func WithLogger(logger Logger) CookieJarOption {
	return func(config *cookieJarConfig) {
		config.logger = logger
	}
}

type CookieJar interface {
	http.CookieJar
	GetAllCookies() map[url.URL][]*http.Cookie
}

type cookieJar struct {
	jar        *cookiejar.Jar
	config     *cookieJarConfig
	allCookies map[url.URL][]*http.Cookie
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

	c := &cookieJar{
		jar:        realJar,
		config:     config,
		allCookies: make(map[url.URL][]*http.Cookie),
	}

	return c

}

func (jar *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Lock()
	defer jar.Unlock()

	if !jar.config.skipExisting {
		existingCookies := jar.allCookies[*u]

		var remainingExistingCookies []*http.Cookie

		for _, existingCookie := range existingCookies {
			shouldOverwrite := false
			for _, cookie := range cookies {
				shouldOverwrite = existingCookie.Name == cookie.Name

				if shouldOverwrite {
					break
				}
			}

			if shouldOverwrite {
				jar.config.logger.Debug("cookie %s should be overwriten by newer value", existingCookie.Name)
				continue
			}

			remainingExistingCookies = append(remainingExistingCookies, existingCookie)
		}

		newCookies := append(remainingExistingCookies, cookies...)

		jar.jar.SetCookies(u, newCookies)
		jar.allCookies[*u] = newCookies

		return
	}

	var filteredCookies []*http.Cookie

	existingCookies := jar.allCookies[*u]

	for _, cookie := range cookies {
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

		filteredCookies = append(filteredCookies, cookie)
	}

	newCookies := append(existingCookies, filteredCookies...)
	jar.jar.SetCookies(u, newCookies)
	jar.allCookies[*u] = newCookies
}

func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	return jar.allCookies[*u]
}

func (jar *cookieJar) GetAllCookies() map[url.URL][]*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()

	copied := make(map[url.URL][]*http.Cookie)
	for u, c := range jar.allCookies {
		copied[u] = c
	}

	return copied
}

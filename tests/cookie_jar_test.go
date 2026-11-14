package tests

import (
	"io"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/bogdanfinn/tls-client/profiles"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/httptest"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/stretchr/testify/assert"
)

func TestClient_SkipExistingCookiesOnClientSetCookies(t *testing.T) {
	jarOptions := []tls_client.CookieJarOption{
		tls_client.WithSkipExisting(),
	}

	jar := tls_client.NewCookieJar(jarOptions...)

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	u := &url.URL{
		Scheme: "http",
		Host:   "testhost.de",
		Path:   "/test",
	}

	assert.Equal(t, 0, len(client.GetCookies(u)))

	cookie := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: 1,
	}

	client.SetCookies(u, []*http.Cookie{cookie})

	assert.Equal(t, 1, len(client.GetCookies(u)))

	cookie2 := &http.Cookie{
		Name:   "test2",
		Value:  "test2",
		MaxAge: 1,
	}
	client.SetCookies(u, []*http.Cookie{cookie2})

	assert.Equal(t, 2, len(client.GetCookies(u)))

	cookie3 := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: 1,
	}
	client.SetCookies(u, []*http.Cookie{cookie3})

	assert.Equal(t, 2, len(client.GetCookies(u)))
}

func TestClient_SkipExistingCookiesOnSetCookiesResponse(t *testing.T) {
	testServer := getCookieWebServer()
	testServer.Start()
	defer testServer.Close()

	jarOptions := []tls_client.CookieJarOption{
		tls_client.WithSkipExisting(),
	}

	jar := tls_client.NewCookieJar(jarOptions...)

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(testServer.URL + "/cookies")
	if err != nil {
		t.Fatal(err)
	}

	doCookieRequest(t, client, u.String())

	cookiesAfterFirstRequest := client.GetCookies(u)

	assert.Equal(t, len(mockServerCookieNames), len(cookiesAfterFirstRequest))
	assertCookieValues(t, cookiesAfterFirstRequest, "1")

	existingCookie := &http.Cookie{
		Name:   cookiesAfterFirstRequest[0].Name,
		Value:  "manual",
		MaxAge: 60,
	}
	client.SetCookies(u, []*http.Cookie{existingCookie})

	cookiesAfterSetCookies := client.GetCookies(u)

	assert.Equal(t, len(mockServerCookieNames), len(cookiesAfterSetCookies))
	assertCookieValues(t, cookiesAfterSetCookies, "1")

	// The server answers the second request with new values for the same names,
	// which the jar has to skip because the names already exist.
	doCookieRequest(t, client, u.String())

	cookiesAfterSecondRequest := client.GetCookies(u)

	assert.Equal(t, len(mockServerCookieNames), len(cookiesAfterSecondRequest))
	assertCookieValues(t, cookiesAfterSecondRequest, "1")
}

var mockServerCookieNames = []string{"session", "consent", "tracking"}

// getCookieWebServer sets the same cookie names on every response, with the
// request count as value, so a changed value shows that a cookie was overwritten.
func getCookieWebServer() *httptest.Server {
	var requestCount atomic.Int64

	var cookieHandler = func(w http.ResponseWriter, req *http.Request) {
		value := strconv.FormatInt(requestCount.Add(1), 10)

		for _, name := range mockServerCookieNames {
			http.SetCookie(w, &http.Cookie{
				Name:   name,
				Value:  value,
				Path:   "/",
				MaxAge: 60,
			})
		}

		w.WriteHeader(http.StatusOK)
	}

	router := http.NewServeMux()
	router.HandleFunc("/cookies", cookieHandler)

	return httptest.NewUnstartedServer(router)
}

func doCookieRequest(t *testing.T, client tls_client.HttpClient, urlString string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, urlString, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func assertCookieValues(t *testing.T, cookies []*http.Cookie, expectedValue string) {
	t.Helper()

	for _, cookie := range cookies {
		assert.Equal(t, expectedValue, cookie.Value, "cookie %s", cookie.Name)
	}
}

func TestClient_ExcludeExpiredCookiesFromRequest(t *testing.T) {
	jar := tls_client.NewCookieJar()

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
	u := &url.URL{
		Scheme: "http",
		Host:   "testhost.de",
		Path:   "/test",
	}

	assert.Equal(t, 0, len(client.GetCookies(u)))

	cookieAlive := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: 1,
	}

	cookieExpired := &http.Cookie{
		Name:   "test2",
		Value:  "test2",
		MaxAge: -1,
	}

	client.SetCookies(u, []*http.Cookie{cookieAlive, cookieExpired})

	assert.Equal(t, 1, len(client.GetCookies(u)))

	cookieExpireExisting := &http.Cookie{
		Name:   "test1",
		Value:  "test1",
		MaxAge: -1,
	}
	client.SetCookies(u, []*http.Cookie{cookieExpireExisting})

	assert.Equal(t, 0, len(client.GetCookies(u)))
}

// TestClient_CookieDomainRules covers issue #75: a cookie set for a domain
// that does not match the request host must never be stored or sent, and a
// host-only cookie set on a subdomain must not leak to the parent domain.
func TestClient_CookieDomainRules(t *testing.T) {
	jar := tls_client.NewCookieJar()

	options := []tls_client.HttpClientOption{
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		t.Fatal(err)
	}

	u := &url.URL{
		Scheme: "https",
		Host:   "httpbin.org",
		Path:   "/cookies",
	}

	client.SetCookies(u, []*http.Cookie{
		{Name: "host", Value: "1"},
		// foreign domain: rejected at store time, like the standard jar does
		{Name: "foreign", Value: "x", Domain: "example.org"},
		// malformed domain: rejected at store time
		{Name: "malformed", Value: "x", Domain: "."},
	})

	cookies := client.GetCookies(u)
	assert.Equal(t, 1, len(cookies))
	assert.Equal(t, "host", cookies[0].Name)

	// the foreign cookie does not appear in the exported view either
	for _, stored := range jar.GetAllCookies() {
		for _, c := range stored {
			assert.NotEqual(t, "foreign", c.Name)
		}
	}

	sub := &url.URL{
		Scheme: "https",
		Host:   "sub.httpbin.org",
		Path:   "/cookies",
	}

	// host-only cookie set on a subdomain: no Domain attribute means it must
	// not go to the parent domain
	client.SetCookies(sub, []*http.Cookie{{Name: "subonly", Value: "1"}})

	parentCookies := client.GetCookies(u)
	for _, c := range parentCookies {
		assert.NotEqual(t, "subonly", c.Name)
	}

	subCookies := client.GetCookies(sub)
	foundSubOnly := false
	for _, c := range subCookies {
		if c.Name == "subonly" {
			foundSubOnly = true
		}
	}
	assert.True(t, foundSubOnly)

	// a subdomain may set a domain cookie for its parent; it is then sent to
	// both the subdomain and the parent
	client.SetCookies(sub, []*http.Cookie{{Name: "shared", Value: "1", Domain: "httpbin.org"}})

	foundShared := false
	for _, c := range client.GetCookies(u) {
		if c.Name == "shared" {
			foundShared = true
		}
	}
	assert.True(t, foundShared)

	foundShared = false
	for _, c := range client.GetCookies(sub) {
		if c.Name == "shared" {
			foundShared = true
		}
	}
	assert.True(t, foundShared)

	// a secure cookie does not travel over plain http
	insecure := &url.URL{
		Scheme: "http",
		Host:   "httpbin.org",
		Path:   "/cookies",
	}

	client.SetCookies(u, []*http.Cookie{{Name: "secure", Value: "1", Secure: true}})

	for _, c := range client.GetCookies(insecure) {
		assert.NotEqual(t, "secure", c.Name)
	}
}

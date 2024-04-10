# TLS-Client

### Preface

This TLS Client is built upon https://github.com/Carcraftz/fhttp and https://github.com/Carcraftz/utls (https://github.com/refraction-networking/utls). Big thanks to
all contributors so far. Sadly it seems that the original repositories from Carcraftz are not maintained anymore.

### What is TLS Fingerprinting?

Some people think it is enough to change the user-agent header of a request to let the server think that the client
requesting a resource is a specific browser.
Nowadays this is not enough, because the server might use a technique to detect the client browser which is called TLS
Fingerprinting.

Even though this article is about TLS Fingerprinting in NodeJS it well describes the technique in general.
https://httptoolkit.tech/blog/tls-fingerprinting-node-js/#how-does-tls-fingerprinting-work

### Why is this library needed?

With this library you are able to create a http client implementing an interface which is similar to golangs net/http
client interface.
This TLS Client allows you to specify the Client (Browser and Version) you want to use, when requesting a server.

The Interface of the HTTP Client looks like the following and extends the base net/http Client Interface by some useful functions.
Most likely you will use the `Do()` function like you did before with net/http Client.
```go
type HttpClient interface {
    GetCookies(u *url.URL) []*http.Cookie
    SetCookies(u *url.URL, cookies []*http.Cookie)
    SetCookieJar(jar http.CookieJar)
    GetCookieJar() http.CookieJar
    SetProxy(proxyUrl string) error
    GetProxy() string
    SetFollowRedirect(followRedirect bool)
    GetFollowRedirect() bool
    CloseIdleConnections()
    Do(req *http.Request) (*http.Response, error)
    Get(url string) (resp *http.Response, err error)
    Head(url string) (resp *http.Response, err error)
    Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}
```


### Quick Usage Example

```go
package main

import (
	"fmt"
	"io"
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func main() {
    jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(profiles.Chrome_120),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"*/*"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":                {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-language",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Println(fmt.Sprintf("status code: %d", resp.StatusCode))

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(readBytes))
}
```

### Detailed Documentation

https://bogdanfinn.gitbook.io/open-source-oasis/

### Questions?

Join my discord support server for free: https://discord.gg/7Ej9eJvHqk
No Support in DMs!


### Appreciate my work?

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/CaptainBarnius)


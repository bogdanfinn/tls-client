# TLS-Client

### Preface

This TLS Client is built upon https://github.com/Carcraftz/fhttp and https://github.com/Carcraftz/utls. Big thanks to
all contributors so far. Sadly it seems that the original repositories are not maintained anymore.

### What is TLS Fingerprinting?

Some people think it is enough to change the user-agent header of a request to let the server think that the client
requesting a resource is a specific browser.
Nowadays this is not enough, because the server might use a technique to detect the client browser which is called TLS
Fingerprinting.

Even tho this article is about TLS Fingerprinting in NodeJS it well describes the technique in general.
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
    SetProxy(proxyUrl string) error
    GetProxy() string
    SetFollowRedirect(followRedirect bool)
    GetFollowRedirect() bool
    Do(req *http.Request) (*http.Response, error)
    Get(url string) (resp *http.Response, err error)
    Head(url string) (resp *http.Response, err error)
    Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}
```

### Detailed Documentation

https://bogdanfinn.gitbook.io/open-source-oasis/

### Questions?

Join my discord support server for free: https://discord.gg/7Ej9eJvHqk
No Support in DMs!
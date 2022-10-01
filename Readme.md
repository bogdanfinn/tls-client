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
    GetCookieJar() http.CookieJar
    GetCookies(u *url.URL) []*http.Cookie
    SetCookies(u *url.URL, cookies []*http.Cookie)
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

### Supported and tested Clients

- Chrome
    - 103 (chrome_103)
    - 104 (chrome_104)
    - 105 (chrome_105)
- Safari
    - 15.6.1 (safari_15_6_1)
    - 16.0 (safari_16_0)
- iOS (Safari)
    - 15.5 (safari_ios_15_5)
    - 15.6 (safari_ios_15_6)
    - 16.0 (safari_ios_16_0)
- iPadOS (Safari)
    - 15.6 (safari_ios_15_6)
- Firefox
    - 102 (firefox_102)
    - 104 (firefox_104)
    - 105 (firefox_105)
- Opera
    - 89 (opera_89)
    - 90 (opera_90)
    - 91 (opera_91)
- Custom Clients
    - Zalando Android Mobile (zalando_android_mobile) 
    - Nike IOS Mobile (nike_ios_mobile)
    - Nike Android Mobile (nike_android_mobile)

You can also provide your own client. See the example how to do it.

#### Need other clients?

Please open an issue on this github repository. In the best case you provide the response of https://tls.peet.ws/api/all requested by the client you want to be implemented.

### Installation

```go
go get -u github.com/bogdanfinn/tls-client

// or specific version:
// go get github.com/bogdanfinn/tls-client@v0.5.2
```

I would recommend to check the github tags for the latest version and install that one explicit.

### Quick Usage Example

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

func main() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
		tls_client.WithNotFollowRedirects(), 
		//tls_client.WithProxyUrl("http://user:pass@host:ip"),
		//tls_client.WithInsecureSkipVerify(), 
		//tls_client.WithCookieJar(cJar), // create cookieJar instance and pass it as argument
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
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`" Not A;Brand";v="99", "Chromium";v="101", "Google Chrome";v="101"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
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

	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(readBytes))
}

```

For more configured clients check `./profiles.go`, `./custom_profiles.go` or use your own custom client. See examples folder.

#### Default Client
The implemented default client is currently Chrome 105 with a configured request timeout of 30 seconds and no automatic redirect following.

### Compile this client for use in Python as a shared library
Please take a look at the cross compile build script in `cffi_dist/build.sh` to build this tls-client as a shared library for other programming languages (.dll, .so, .dylib).

The build script is written to cross compile from OSX to all other platforms (osx, linux, windows). If your build os is not OSX you might need to adjust the build script.

You can also use the prebuilt packages in `cffi_dist/dist`

A python example on how to load and call the functionality can be found in `cffi_dist/example_python/example.py`. Please be aware that i'm not a python expert.
For more documentation please take a look at: https://github.com/bogdanfinn/tls-client-api

Build and tested with python 3.8 on MacOS.

I highly recommend to take a look at this repository, when you want to use this tls-client in python: https://github.com/FlorianREGAZ/Python-Tls-Client

### Compile this client for use in NodeJS as a shared library
Please take a look at the cross compile build script in `cffi_dist/build.sh` to build this tls-client as a shared library for other programming languages (.dll, .so, .dylib).

The build script is written to cross compile from OSX to all other platforms (osx, linux, windows). If your build os is not OSX you might need to adjust the build script.

You can also use the prebuilt packages in `cffi_dist/dist`

A NodeJS example on how to load and call the functionality can be found in `cffi_dist/example_node/index.js`. Please be aware that you need to run `npm install` to install the node dependencies.
For more documentation please take a look at: https://github.com/bogdanfinn/tls-client-api

Build and tested with nodejs v16.13.2 on MacOS.

### Further Information

This library uses the following api: https://tls.peet.ws/api/all to verify the hashes and fingerprints for akamai and
ja3.

If you are not using go and do not want to implement the shared library but want to use the functionality check out this repository https://github.com/bogdanfinn/tls-client-api

### Frequently Asked Questions / Errors
* **I'm receiving `tls: error decoding message` when using this TLS Client.**

This issue should be fixed since `v.0.3.0`. There was an issue with the CompressCertExtension in the utls package dependency.

* **The TLS-Client does not set the user-agent header correctly**

Do not mix up TLS-Fingerprints with HTTP Request Headers. They have more or less nothing in common. AntiBots using for example header order in addition to TLS-Fingerprinting. This library does only handle the TLS- and Akamai Fingerprint. You are still responsible to define the to be used headers and the header order.
If you do not provide any headers the http client will use by default these two headers (nothing more):
```
accept-encoding: gzip, deflate, br
user-agent: Go-http-client/2.0
```

* **If I use the shared library in electron my application freezes?**
Please only load the dll once in your application and call every function `async` to not block the main thread.

* **My Post Request is not working correctly?**
Please make sure that you set the correct `Content-Type` Header for your Post Body Payload.

### Questions?

Join my discord support server: https: // discord.gg / 7Ej9eJvHqk 
No Support in DMs!
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


### Compile this client as a shared library for use in other languages like Python or NodeJS
Please take a look at the cross compile build script in `cffi_dist/build.sh` to build this tls-client as a shared library for other programming languages (.dll, .so, .dylib).

The build script is written to cross compile from OSX to all other platforms (osx, linux, windows). If your build os is not OSX you might need to adjust the build script.

You can also use the prebuilt packages in `cffi_dist/dist`

A python example on how to load and call the functionality can be found in `cffi_dist/example_python`. Please be aware that i'm not a python expert.
I highly recommend to take a look at this repository, when you want to use this tls-client in python: https://github.com/FlorianREGAZ/Python-Tls-Client

A NodeJS example on how to load and call the functionality can be found in `cffi_dist/example_node`. Please be aware that you need to run `npm install` to install the node dependencies.

The basic logic behind the shared library is, that you pass all required information in a JSON string to the shared lib function which then creates the client, the request and the request data out of it and forwards the request.
For more documentation on this JSON string please take a look at: https://github.com/bogdanfinn/tls-client-api

Every Response from the shared library will contain an id field like that: `"id":"some-uuid-v4-value"`. You can use the id to free the memory. Otherwise you will end up with growing allocated memory.
Basic Nodejs example:
```js
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));
const responseObject = JSON.parse(response)
tlsClientLibrary.freeMemory(responseObject.Id)
```

### Frequently Asked Questions / Errors
* **How can I add `GREASE` to Custom Client Profiles when using the shared library?**

Please refer to `index_custom_client.js` or `example_custom_client.py` in either NodeJS examples or Python Examples. You can use the Magic Number `2570`. You can add it in a ja3 string for example to turn that into a `GREASE` cipher or tls extension.

* **I receive PROTOCOL_ERROR on POST Request**

This is a very generic error and can have many root causes. Most likely users of this client are setting the `content-length` header manually with a (wrong) fixed value.

* **This client fails when I test on www.google.com**

Please check this issue for explanation: https://github.com/bogdanfinn/tls-client/issues/6. Should be fixed since 0.8.3

* **I'm receiving `tls: error decoding message` when using this TLS Client.**

This issue should be fixed since `v0.3.0`. There was an issue with the CompressCertExtension in the utls package dependency.

* **The TLS-Client does not set the user-agent header correctly**

Do not mix up TLS-Fingerprints with HTTP Request Headers. They have more or less nothing in common. AntiBots using for example header order in addition to TLS-Fingerprinting. This library does only handle the TLS- and Akamai Fingerprint. You are still responsible to define the to be used headers and the header order.
If you do not provide any headers the http client will use by default these two headers (nothing more):
```
accept-encoding: gzip, deflate, br
user-agent: Go-http-client/2.0
```

* **If I use the shared library in electron my application freezes?**
Please only load the dll once in your application and call every function `async` to not block the main thread. An example is added in the nodejs examples.

* **My Post Request is not working correctly?**
Please make sure that you set the correct `Content-Type` Header for your Post Body Payload.

* **About `accept-encoding` and automatic decompression**
If you are specifying `accept-encoding` header yourself and you are on `http1` connection than you have to take care of the **decompression yourself**.
It is not done automatically. Only if you are not adding `accept-encoding` header then the library adds it for you if not explicit disabled and also handles the decompression automatically.

On `http2` the automatic decompression should always be in place according to the `Content-Type` Header on the Response.

So if you are trying to use "only" `accept-encoding: gzip` you have to take care of the decompression.
`DecompressBody` is exported. You can just reuse it like that:

```go
req, err := http.NewRequest(http.MethodGet, "https://tls.browserleaks.com/json", nil)
    if err != nil {
        log.Println(err)
        return
    }

    req.Header = http.Header{
        "accept":          {"*/*"},
        "accept-encoding": {"gzip"},
        "accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
        "user-agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
        http.HeaderOrderKey: {
            "accept",
            "accept-encoding",
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

    decomBody := http.DecompressBody(resp)

    all, err := io.ReadAll(decomBody)
    if err != nil {
        log.Println(err)
        return
    }
    log.Println(string(all))
```

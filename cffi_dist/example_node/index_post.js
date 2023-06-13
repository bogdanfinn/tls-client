const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64-1.4.0.dylib', {
    'request': ['string', ['string']],
    'getCookiesFromSession': ['string', ['string']],
    'addCookiesToSession': ['string', ['string']],
    'freeMemory': ["void", ['string']],
    'destroyAll': ['string', []],
    'destroySession': ['string', ['string']]
});

const requestPayload = {
    "tlsClientIdentifier": "chrome_103",
    "followRedirects": false,
    "insecureSkipVerify": false,
    "withoutCookieJar": false,
    "withDefaultCookieJar": false,
    "forceHttp1": false,
    "withDebug": false,
    "withRandomTLSExtensionOrder": false,
    "isByteResponse": false,
    "isByteRequest": false,
    "catchPanics": false,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "proxyUrl": "",
    "isRotatingProxy": false,
    "certificatePinningHosts": {},
    "headers": {
        "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
        "accept-encoding": "gzip, deflate, br",
        "content-type": "application/x-www-form-urlencoded",
        "accept-language": "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"
    },
    "headerOrder": [
        "accept",
        "user-agent",
        "content-type",
        "accept-encoding",
        "accept-language"
    ],
    "requestUrl": "https://eonk4gg5hquk0g6.m.pipedream.net",
    "requestMethod": "POST",
    "requestBody": "foo=bar&baz=foo",
    "requestCookies": []
}

// call the library with the requestPayload as string
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));

// convert response string to json
const responseObject = JSON.parse(response)

console.log(responseObject)
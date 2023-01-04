const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64-1.3.0dylib', {
    'request': ['string', ['string']],
    'getCookiesFromSession': ['string', ['string']],
    'addCookiesToSession': ['string', ['string']],
    'freeMemory': ["void", ['string']],
    'destroyAll': ['string', []],
    'destroySession': ['string', ['string']]
});

const requestPayload = {
    "tlsClientIdentifier": "chrome_108",
    "followRedirects": true,
    "insecureSkipVerify": false,
    "withoutCookieJar": false,
    "withDefaultCookieJar": false,
    "isByteRequest": false,
    "withDebug": false,
    "forceHttp1": false,
    "withRandomTLSExtensionOrder": true,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "certificatePinningHosts": {},
    "sessionId": "asos",
    "proxyUrl": "",
    "headers": {
        "accept-encoding": "gzip, deflate, br",
        "accept-language": "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7",
        "accept": "*/*",
        "cache-control": "no-cache",
        "sec-ch-ua": `"Google Chrome";v="107", "Chromium";v="107", "Not=A?Brand";v="24"`,
        "sec-ch-ua-mobile": "?0",
        "sec-ch-ua-platform": `"macOS"`,
        "sec-fetch-dest": "empty",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"
    },
    "headerOrder": [
        "accept",
        "accept-encoding",
        "accept-language",
        "cache-control",
        "sec-ch-ua",
        "sec-ch-ua-mobile",
        "sec-ch-ua-platform",
        "sec-fetch-dest",
        "user-agent",
    ],
    "requestUrl": "https://my.asos.com",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

// call the library with the requestPayload as string
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));


// convert response string to json
const responseObject = JSON.parse(response)

console.log("status", responseObject.status)
console.log("target", responseObject.target)
tlsClientLibrary.freeMemory(responseObject.Id)

const payload = {
    sessionId: 'asos',
    url: "https://my.asos.com",
}

const cookiesResponse = tlsClientLibrary.getCookiesFromSession(JSON.stringify(payload))

const cookiesInSession = JSON.parse(cookiesResponse)

cookiesInSession.cookies.map(c => console.log(c.Name, c.Value, c.Domain, c.Path))
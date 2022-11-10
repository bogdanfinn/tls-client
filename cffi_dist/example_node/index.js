const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64-0.8.3.dylib', {
    'request': ['string', ['string']],
    'getCookiesFromSession': ['string', ['string']],
    'freeAll': ['string', []],
    'freeSession': ['string', ['string']]
});

const requestPayload = {
    "tlsClientIdentifier": "chrome_103",
    "followRedirects": false,
    "insecureSkipVerify": false,
    "withoutCookieJar": false,
    "timeoutSeconds": 30,
    "sessionId": "my-session-id",
    "proxyUrl": "",
    "headers": {
        "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36",
        "accept-encoding": "gzip, deflate, br",
        "accept-language": "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"
    },
    "headerOrder": [
        "accept",
        "user-agent",
        "accept-encoding",
        "accept-language"
    ],
    "requestUrl": "https://microsoft.com",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

// call the library with the requestPayload as string
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));

// convert response string to json
const responseObject = JSON.parse(response)

console.log(responseObject)

const payload = {
    sessionId: 'my-session-id',
    url: "https://example.com",
}

const cookiesResponse = tlsClientLibrary.getCookiesFromSession(JSON.stringify(payload))

const cookiesInSession = JSON.parse(cookiesResponse)

console.log(cookiesInSession)


const freeSessionPayload = {
    sessionId: 'my-session-id',
}

const freeSessionResponse = tlsClientLibrary.freeSession(JSON.stringify(freeSessionPayload))

const freeSessionResponseParsed = JSON.parse(freeSessionResponse)

console.log(freeSessionResponseParsed)
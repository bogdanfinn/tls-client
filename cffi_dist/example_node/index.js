const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64-1.7.2.dylib', {
    'request': ['string', ['string']],
    'getCookiesFromSession': ['string', ['string']],
    'addCookiesToSession': ['string', ['string']],
    'freeMemory': ["void", ['string']],
    'destroyAll': ['string', []],
    'destroySession': ['string', ['string']]
});

const requestPayload = {
    "tlsClientIdentifier": "chrome_103",
    "followRedirects": true,
    "insecureSkipVerify": false,
    "withoutCookieJar": false,
    "withDefaultCookieJar": false,
    "isByteRequest": false,
    "catchPanics": false,
    "withDebug": false,
    "forceHttp1": false,
    "withRandomTLSExtensionOrder": false,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "sessionId": "my-session-id",
    "proxyUrl": "",
    "isRotatingProxy": false,
    "certificatePinningHosts": {},
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
tlsClientLibrary.freeMemory(responseObject.id)

const payload = {
    sessionId: 'my-session-id',
    url: "https://microsoft.com",
}

const cookiesResponse = tlsClientLibrary.getCookiesFromSession(JSON.stringify(payload))

const cookiesInSession = JSON.parse(cookiesResponse)

console.log(cookiesInSession)


const destroySessionPayload = {
    sessionId: 'my-session-id',
}

const destroySessionResponse = tlsClientLibrary.destroySession(JSON.stringify(destroySessionPayload))

const destroySessionResponseParsed = JSON.parse(destroySessionResponse)

console.log(destroySessionResponseParsed)
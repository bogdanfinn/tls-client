const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-xgo-1.16.0-darwin-amd64.dylib', {
    'request': ['string', ['string']],
    'freeMemory': ['void', ['string']],
    'destroySession': ['string', ['string']]
});

const requestPayload = {
    "tlsClientIdentifier": "chrome_150",
    "followRedirects": false,
    "insecureSkipVerify": false,
    "withoutCookieJar": false,
    "withCustomCookieJar": false,
    "isByteRequest": false,
    "forceHttp1": false,
    "withDebug": false,
    "catchPanics": false,
    "withRandomTLSExtensionOrder": false,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "sessionId": "my-redirect-session",
    "proxyUrl": "",
    "isRotatingProxy": false,
    "certificatePinningHosts": {},
    "headers": {
        "accept": "*/*",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36",
    },
    "headerOrder": ["accept", "user-agent"],
    "requestUrl": "https://httpbin.org/redirect/1",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

let response = JSON.parse(tlsClientLibrary.request(JSON.stringify(requestPayload)));
console.log('followRedirects=false:', response.status, response.target);

// Every response allocates memory on the go side. Free it with its "id"
// as soon as you are done reading it, otherwise the memory is never released.
tlsClientLibrary.freeMemory(response.id);

// followRedirects can be changed within an existing session.
requestPayload.followRedirects = true;

response = JSON.parse(tlsClientLibrary.request(JSON.stringify(requestPayload)));
console.log('followRedirects=true:', response.status, response.target);
tlsClientLibrary.freeMemory(response.id);

tlsClientLibrary.destroySession(JSON.stringify({ sessionId: 'my-redirect-session' }));

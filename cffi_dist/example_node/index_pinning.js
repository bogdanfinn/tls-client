const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-xgo-1.16.0-darwin-amd64.dylib', {
    'request': ['string', ['string']],
    'freeMemory': ['void', ['string']]
});

// Pins are the base64 encoded SHA-256 hashes of a host's public keys.
// Generate them with: hpkp-pins -server=bstn.com:443
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
    "sessionId": null,
    "proxyUrl": "",
    "isRotatingProxy": false,
    "certificatePinningHosts": {
        "bstn.com": [
            "NQvy9sFS99nBqk/nZCUF44hFhshrkvxqYtfrZq3i+Ww=",
            "4a6cPehI7OG6cuDZka5NDZ7FR8a60d3auda+sKfg4Ng=",
            "x4QzPSC810K5/cMjb05Qm4k3Bw5zBn4lTdO/nEW/Td4=",
        ]
    },
    "headers": {
        "accept": "*/*",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36",
    },
    "headerOrder": ["accept", "user-agent"],
    "requestUrl": "https://bstn.com",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

const response = JSON.parse(tlsClientLibrary.request(JSON.stringify(requestPayload)));

// If a pin does not match, "status" is 0 and "body" holds the pinning error
// instead of a remote response.
console.log(response);

// Every response allocates memory on the go side. Free it with its "id"
// as soon as you are done reading it, otherwise the memory is never released.
tlsClientLibrary.freeMemory(response.id);

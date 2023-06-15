const ffi = require('ffi-napi');
const fs = require("fs")

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64-1.4.0.dylib', {
    'request': ['string', ['string']],
    'getCookiesFromSession': ['string', ['string']],
    'addCookiesToSession': ['string', ['string']],
    'freeMemory': ["void", ['string']],
    'destroyAll': ['string', []],
    'destroySession': ['string', ['string']]
});

const contents = fs.readFileSync('./cb_example.png', {encoding: 'base64'});

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
    "isByteRequest": true,
    "catchPanics": false,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "certificatePinningHosts": {},
    "proxyUrl": "",
    "isRotatingProxy": false,
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
    "requestUrl": "https://eo1wmj45078deme.m.pipedream.net",
    "requestMethod": "POST",
    "requestBody": contents,
    "requestCookies": []
}

// call the library with the requestPayload as string
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));

// convert response string to json
const responseObject = JSON.parse(response)

console.log(responseObject)
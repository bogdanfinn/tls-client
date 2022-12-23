const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64-1.1.2.dylib', {
    'request': ['string', ['string']],
    'getCookiesFromSession': ['string', ['string']],
    'freeMemory': ["void", ['string']],
    'destroyAll': ['string', []],
    'destroySession': ['string', ['string']]
});

const requestPayload = {
    "tlsClientIdentifier": "chrome_107",
    "followRedirects": false,
    "insecureSkipVerify": false,
    "sessionId": "footlocker",
    "withoutCookieJar": false,
    "isByteRequest": false,
    "withDebug": true,
    "forceHttp1": false,
    "withRandomTLSExtensionOrder": false,
    "timeoutSeconds": 30,
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
    "requestUrl": "https://www.footlocker.de/",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

// call the library with the requestPayload as string
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));

// convert response string to json
const responseObject = JSON.parse(response)

console.log("Cookies on Response:", responseObject.cookies)

const payload = {
    sessionId: 'footlocker',
    url: "https://www.footlocker.de/",
}

const cookiesResponse = tlsClientLibrary.getCookiesFromSession(JSON.stringify(payload))

const cookiesInSession = JSON.parse(cookiesResponse)

cookiesInSession.map(cookieInSession => {
    console.log("cookie in session: ", cookieInSession.Name, cookieInSession.Value)
})


const requestPayload2 = {
    "tlsClientIdentifier": "chrome_107",
    "followRedirects": false,
    "insecureSkipVerify": false,
    "sessionId": "footlocker",
    "withoutCookieJar": false,
    "withDebug": true,
    "isByteRequest": false,
    "forceHttp1": false,
    "withRandomTLSExtensionOrder": false,
    "timeoutSeconds": 30,
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
    "requestUrl": "https://www.footlocker.de/",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": [{
        "name": "datadome",
        "value": "overwrittenValue",
        //Domain: ".footlocker.de",
        //Path:   "/",
    }]
}

// call the library with the requestPayload as string
const response2 = tlsClientLibrary.request(JSON.stringify(requestPayload2));

// convert response string to json
const responseObject2 = JSON.parse(response2)

console.log("Cookies on Response:", responseObject2.cookies)

const payload2 = {
    sessionId: 'footlocker',
    url: "https://www.footlocker.de/",
}

const cookiesResponse2 = tlsClientLibrary.getCookiesFromSession(JSON.stringify(payload2))

const cookiesInSession2 = JSON.parse(cookiesResponse2)

cookiesInSession2.map(cookieInSession => {
    console.log("cookie in session: ", cookieInSession.Name, cookieInSession.Value)
})

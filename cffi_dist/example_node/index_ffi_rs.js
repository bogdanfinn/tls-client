const { load, DataType, open } = require('ffi-rs')

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const LIBRARY_NAME = "tls-client";
open({
    library: LIBRARY_NAME,
    path: "./../dist/tls-client-xgo-1.7.2-linux-amd64.so"
});

function request(requestData) {
    return JSON.parse(load({
        library: LIBRARY_NAME,
        funcName: "request",
        retType: DataType.String,
        paramsType: [DataType.String],
        paramsValue: [JSON.stringify(requestData)]
    }));
}

function freeMemory(id) {
    load({
        library: LIBRARY_NAME,
        funcName: "freeMemory",
        retType: DataType.Void,
        paramsType: [DataType.String],
        paramsValue: [id]
    });
}

function getCookiesFromSession(requestData) {
    return JSON.parse(load({
        library: LIBRARY_NAME,
        funcName: "getCookiesFromSession",
        retType: DataType.String,
        paramsType: [DataType.String],
        paramsValue: [JSON.stringify(requestData)]
    }));
}

function addCookiesToSession(cookieData) {
    return JSON.parse(load({
        library: LIBRARY_NAME,
        funcName: "addCookiesToSession",
        retType: DataType.String,
        paramsType: [DataType.String],
        paramsValue: [JSON.stringify(cookieData)]
    }));
}

function destroySession(sessionId) {
    return JSON.parse(load({
        library: LIBRARY_NAME,
        funcName: "destroySession",
        retType: DataType.String,
        paramsType: [DataType.String],
        paramsValue: [JSON.stringify({sessionId})]
    }));
}

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
const response = request(requestPayload);
console.log(response)
freeMemory(response.id)

const payload = {
    sessionId: 'my-session-id',
    url: "https://microsoft.com",
}

const cookiesResponse = getCookiesFromSession(payload)
console.log(cookiesResponse)

const destroySessionResponse = destroySession('my-session-id')
console.log(destroySessionResponse)
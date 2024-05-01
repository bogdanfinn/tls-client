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
    "followRedirects": false,
    "insecureSkipVerify": false,
    "withoutCookieJar": false,
    "withDefaultCookieJar": false,
    "isByteRequest": false,
    "catchPanics": false,
    "forceHttp1": false,
    "withDebug": false,
    "withRandomTLSExtensionOrder": false,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "sessionId": "my-session-id",
    "certificatePinningHosts": {},
    "customTlsClient": {
        "ja3String": "771,2570-4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,2570-0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-2570-21,2570-29-23-24,0",
        "h2Settings": {
            "HEADER_TABLE_SIZE": 65536,
            "MAX_CONCURRENT_STREAMS": 1000,
            "INITIAL_WINDOW_SIZE": 6291456,
            "MAX_HEADER_LIST_SIZE": 262144
        },
        "h2SettingsOrder": [
            "HEADER_TABLE_SIZE",
            "MAX_CONCURRENT_STREAMS",
            "INITIAL_WINDOW_SIZE",
            "MAX_HEADER_LIST_SIZE"
        ],
        "supportedSignatureAlgorithms": [
            "ECDSAWithP256AndSHA256",
            "PSSWithSHA256",
            "PKCS1WithSHA256",
            "ECDSAWithP384AndSHA384",
            "PSSWithSHA384",
            "PKCS1WithSHA384",
            "PSSWithSHA512",
            "PKCS1WithSHA512",
        ],
        "supportedVersions": ["GREASE", "1.3", "1.2"],
        "keyShareCurves": ["GREASE", "X25519"],
        "certCompressionAlgo": "brotli",
        "alpnProtocols": ["h2", "http/1.1"],
        "alpsProtocols": ["h2"],
        "pseudoHeaderOrder": [
            ":method",
            ":authority",
            ":scheme",
            ":path"
        ],
        "connectionFlow": 15663105,
        "priorityFrames": [{
            "streamID": 1,
            "priorityParam": {
                "streamDep": 1,
                "exclusive": true,
                "weight": 1
            }
        }],
        "headerPriority": {
            "streamDep": 1,
            "exclusive": true,
            "weight": 1
        },
    },
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
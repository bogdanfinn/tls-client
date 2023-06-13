import ctypes
import json

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64-1.4.0.dylib')

# extract the exposed request function from the shared package
request = library.request
request.argtypes = [ctypes.c_char_p]
request.restype = ctypes.c_char_p

getCookiesFromSession = library.getCookiesFromSession
getCookiesFromSession.argtypes = [ctypes.c_char_p]
getCookiesFromSession.restype = ctypes.c_char_p

addCookiesToSession = library.addCookiesToSession
addCookiesToSession.argtypes = [ctypes.c_char_p]
addCookiesToSession.restype = ctypes.c_char_p

freeMemory = library.freeMemory
freeMemory.argtypes = [ctypes.c_char_p]

destroySession = library.destroySession
destroySession.argtypes = [ctypes.c_char_p]
destroySession.restype = ctypes.c_char_p

destroyAll = library.destroyAll
destroyAll.restype = ctypes.c_char_p

requestPayload = {
    "followRedirects": False,
    "insecureSkipVerify": False,
    "withoutCookieJar": False,
    "withDefaultCookieJar": False,
    "isByteRequest": False,
    "forceHttp1": False,
    "catchPanics": False,
    "withDebug": False,
    "withRandomTLSExtensionOrder": False,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "sessionId": "2my-session-id",
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
        "pseudoHeaderOrder": [
            ":method",
            ":authority",
            ":scheme",
            ":path"
        ],
        "connectionFlow": 15663105,
        "priorityFrames": [],
        "headerPriority": None,
    },
    "proxyUrl": "",
    "isRotatingProxy": False,
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

# this is a pointer to the response
response = request(json.dumps(requestPayload).encode('utf-8'))

# we dereference the pointer to a byte array
response_bytes = ctypes.string_at(response)

# convert our byte array to a string (tls client returns json)
response_string = response_bytes.decode('utf-8')

# convert response string to json
response_object = json.loads(response_string)

# print out output
print(response_object)

cookiePayload = {
    "sessionId": "my-session-id",
    "url": "https://example.com",
}

cookieResponse = getCookiesFromSession(json.dumps(requestPayload).encode('utf-8'))
# we dereference the pointer to a byte array
cookieResponse_bytes = ctypes.string_at(cookieResponse)
# convert our byte array to a string (tls client returns json)
cookieResponse_string = cookieResponse_bytes.decode('utf-8')
# convert response string to json
cookieResponse_object = json.loads(cookieResponse_string)

# print out output
print(cookieResponse_object)

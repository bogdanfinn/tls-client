import ctypes
import json

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-xgo-1.16.0-darwin-amd64.dylib')

# extract the exposed request function from the shared package
request = library.request
request.argtypes = [ctypes.c_char_p]
request.restype = ctypes.c_char_p

freeMemory = library.freeMemory
freeMemory.argtypes = [ctypes.c_char_p]

# Pins are the base64 encoded SHA-256 hashes of a host's public keys.
# Generate them with: hpkp-pins -server=bstn.com:443
requestPayload = {
    "tlsClientIdentifier": "chrome_150",
    "followRedirects": False,
    "insecureSkipVerify": False,
    "withoutCookieJar": False,
    "withCustomCookieJar": False,
    "isByteRequest": False,
    "forceHttp1": False,
    "withDebug": False,
    "catchPanics": False,
    "withRandomTLSExtensionOrder": False,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "sessionId": None,
    "proxyUrl": "",
    "isRotatingProxy": False,
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

response = request(json.dumps(requestPayload).encode('utf-8'))
response_object = json.loads(ctypes.string_at(response).decode('utf-8'))

# If a pin does not match, "status" is 0 and "body" holds the pinning error
# instead of a remote response.
print(response_object)

# Every response allocates memory on the go side. Free it with its "id" as soon
# as you are done reading it, otherwise the memory is never released.
freeMemory(response_object["id"].encode('utf-8'))

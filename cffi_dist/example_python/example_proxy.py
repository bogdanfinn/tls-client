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

destroySession = library.destroySession
destroySession.argtypes = [ctypes.c_char_p]
destroySession.restype = ctypes.c_char_p

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
    "sessionId": "my-proxy-session",
    # http://, socks5:// and socks5h:// proxy URLs are all supported.
    "proxyUrl": "http://user:pass@proxy-one.example.com:8000",
    "isRotatingProxy": False,
    "certificatePinningHosts": {},
    "headers": {
        "accept": "*/*",
        "user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36",
    },
    "headerOrder": ["accept", "user-agent"],
    "requestUrl": "https://tls.peet.ws/api/all",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

response = request(json.dumps(requestPayload).encode('utf-8'))
response_object = json.loads(ctypes.string_at(response).decode('utf-8'))
print("via proxy one:", response_object.get("status"))

# Every response allocates memory on the go side. Free it with its "id" as soon
# as you are done reading it, otherwise the memory is never released.
freeMemory(response_object["id"].encode('utf-8'))

# Swap the proxy for the same session - no new client/session needs to be built.
requestPayload["proxyUrl"] = "http://user:pass@proxy-two.example.com:8000"

response = request(json.dumps(requestPayload).encode('utf-8'))
response_object = json.loads(ctypes.string_at(response).decode('utf-8'))
print("via proxy two:", response_object.get("status"))
freeMemory(response_object["id"].encode('utf-8'))

destroySession(json.dumps({"sessionId": "my-proxy-session"}).encode('utf-8'))

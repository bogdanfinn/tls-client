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
    "tlsClientIdentifier": "chrome_105",
    "followRedirects": False,
    "insecureSkipVerify": False,
    "withoutCookieJar": False,
    "withDefaultCookieJar": False,
    "isByteRequest": False,
    "forceHttp1": False,
    "withDebug": False,
    "catchPanics": False,
    "withRandomTLSExtensionOrder": False,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "sessionId": "my-session-id",
    "proxyUrl": "",
    "isRotatingProxy": False,
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
    "url": "https://microsoft.com",
}

cookieResponse = getCookiesFromSession(json.dumps(cookiePayload).encode('utf-8'))
# we dereference the pointer to a byte array
cookieResponse_bytes = ctypes.string_at(cookieResponse)
# convert our byte array to a string (tls client returns json)
cookieResponse_string = cookieResponse_bytes.decode('utf-8')
# convert response string to json
cookieResponse_object = json.loads(cookieResponse_string)

# print out output
print(cookieResponse_object)


destroySessionPayload = {
    "sessionId": "my-session-id",
}

destroySessionResponse = destroySession(json.dumps(destroySessionPayload).encode('utf-8'))
# we dereference the pointer to a byte array
destroySessionResponse_bytes = ctypes.string_at(destroySessionResponse)
# convert our byte array to a string (tls client returns json)
destroySessionResponse_string = destroySessionResponse_bytes.decode('utf-8')
# convert response string to json
destroySessionResponse_object = json.loads(destroySessionResponse_string)

# print out output
print(destroySessionResponse_object)
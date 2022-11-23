import ctypes
import json

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64-0.9.1.dylib')

# extract the exposed request function from the shared package
request = library.request
request.argtypes = [ctypes.c_char_p]
request.restype = ctypes.c_char_p

getCookiesFromSession = library.getCookiesFromSession
getCookiesFromSession.argtypes = [ctypes.c_char_p]
getCookiesFromSession.restype = ctypes.c_char_p

freeSession = library.freeSession
freeSession.argtypes = [ctypes.c_char_p]
freeSession.restype = ctypes.c_char_p

freeAll = library.freeAll
freeAll.restype = ctypes.c_char_p

requestPayload = {
    "tlsClientIdentifier": "chrome_105",
    "followRedirects": False,
    "insecureSkipVerify": False,
    "withoutCookieJar": False,
    "isByteRequest": False,
    "withRandomTLSExtensionOrder": False,
    "timeoutSeconds": 30,
    "sessionId": "my-session-id",
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


freeSessionPayload = {
    "sessionId": "my-session-id",
}

freeSessionResponse = freeSession(json.dumps(freeSessionPayload).encode('utf-8'))
# we dereference the pointer to a byte array
freeSessionResponse_bytes = ctypes.string_at(freeSessionResponse)
# convert our byte array to a string (tls client returns json)
freeSessionResponse_string = freeSessionResponse_bytes.decode('utf-8')
# convert response string to json
freeSessionResponse_object = json.loads(freeSessionResponse_string)

# print out output
print(freeSessionResponse_object)
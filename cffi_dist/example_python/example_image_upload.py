import ctypes
import json
import base64

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64-1.4.0.dylib')

# extract the exposed request function from the shared package
request = library.request
request.argtypes = [ctypes.c_char_p]
request.restype = ctypes.c_char_p

freeMemory = library.freeMemory
freeMemory.argtypes = [ctypes.c_char_p]

getCookiesFromSession = library.getCookiesFromSession
getCookiesFromSession.argtypes = [ctypes.c_char_p]
getCookiesFromSession.restype = ctypes.c_char_p

addCookiesToSession = library.addCookiesToSession
addCookiesToSession.argtypes = [ctypes.c_char_p]
addCookiesToSession.restype = ctypes.c_char_p

destroySession = library.destroySession
destroySession.argtypes = [ctypes.c_char_p]
destroySession.restype = ctypes.c_char_p

destroyAll = library.destroyAll
destroyAll.restype = ctypes.c_char_p


with open("cb_example.png", "rb") as image_file:
    encoded_string = base64.b64encode(image_file.read())

requestPayload = {
    "tlsClientIdentifier": "chrome_105",
    "followRedirects": False,
    "insecureSkipVerify": False,
    "withoutCookieJar": False,
    "withDefaultCookieJar": False,
    "isByteRequest": True,
    "forceHttp1": False,
    "withDebug": False,
    "catchPanics": False,
    "withRandomTLSExtensionOrder": False,
    "isByteResponse": False,
    "timeoutSeconds": 30,
    "timeoutMilliseconds": 0,
    "certificatePinningHosts": {},
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
    "requestUrl": "https://eo1wmj45078deme.m.pipedream.net",
    "requestMethod": "POST",
    "requestBody": encoded_string.decode('utf-8'),
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

print(response_object)
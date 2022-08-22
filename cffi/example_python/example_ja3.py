import ctypes
import json

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64.dylib')

# extract the exposed request function from the shared package
request = library.request
request.argtypes = [ctypes.c_char_p]
request.restype = ctypes.c_char_p

# build the payload which is needed for the shared package
""" full payload example
{
  "sessionId": "reusableSessionId",
  "tlsClientIdentifier": "chrome_103",
  "followRedirects": False,
  "timeoutSeconds": 30,
  "ja3String": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
  "proxyUrl": "",
  "headerOrder": [
    "key1",
    "key2"
  ],
  "headers": {
    "key1": "value1",
    "key2": "value2"
  },
  "requestCookies": [
    {
      "name": "cookieName",
      "value": "cookieValue",
      "path": "cookiePath",
      "domain": "cookieDomain",
      "expires": "cookieExpires"
    }
  ],
  "requestUrl": "https://tls.peet.ws/api/all",
  "requestBody": "",
  "requestMethod": "GET"
}
"""
requestPayload = {
    "ja3String": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
    "proxyUrl": "",
    "followRedirects": False,
    "timeoutSeconds": 30,
    "headers": {},
    "headerOrder": [],
    "requestUrl": "https://tls.peet.ws/api/all",
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
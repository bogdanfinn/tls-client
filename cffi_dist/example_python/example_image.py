import ctypes
import json
import base64
import re

# load the tls-client shared package for your OS you are currently running your python script (i'm running on mac)
library = ctypes.cdll.LoadLibrary('./../dist/tls-client-darwin-amd64-0.6.1.dylib')

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
  "insecureSkipVerify": False,
  "isByteResponse": true,
  "timeoutSeconds": 30,
  "customTlsClient": {
    "ja3String": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
    "h2Settings": {
      1: 65536,
      3: 1000,
      4: 6291456,
      6: 262144
    },
    "h2SettingsOrder": [
      1,
      3,
      4,
      6
    ],
    "pseudoHeaderOrder": [
      ":method",
      ":authority",
      ":scheme",
      ":path"
    ],
    "connectionFlow": 15663105,
    "priorityFrames": [
      
    ]
  },
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
  "requestBody": "", // needs to be a string! so json.dumps(yourActualyRequestBody) here
  "requestMethod": "GET"
}

The Response from the library looks like that:
{
  "sessionId": "some reusable sessionId",
  "status": 200, //In case of an error the status code will be 0
  "body": "The Response as string here or the error message",
  "headers": {},
  "cookies": {}
}
"""
requestPayload = {
    "tlsClientIdentifier": "chrome_105",
    "followRedirects": False,
    "insecureSkipVerify": False,
    "isByteResponse": True,
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
    "requestUrl": "https://avatars.githubusercontent.com/u/17678241?v=4",
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

data = response_object['body']

dataWithoutMimeType = data.split(",")[1]

with open("./example.png", "wb") as fh:
    fh.write(base64.urlsafe_b64decode(dataWithoutMimeType))
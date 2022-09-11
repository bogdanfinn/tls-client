const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64.dylib', {
    'request': ['string', ['string']]
});

// build the payload which is needed for the shared package
/* full payload example
{
  "sessionId": "reusableSessionId",
  "tlsClientIdentifier": "chrome_103",
  "followRedirects": False,
  "insecureSkipVerify": False,
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
*/
const requestPayload = {
    "tlsClientIdentifier": "chrome_103",
    "followRedirects": false,
    "insecureSkipVerify": false,
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
    "requestUrl": "https://tls.peet.ws/api/all",
    "requestMethod": "GET",
    "requestBody": "",
    "requestCookies": []
}

// call the library with the requestPayload as string
const response = tlsClientLibrary.request(JSON.stringify(requestPayload));

// convert response string to json
const responseObject = JSON.parse(response)

console.log(responseObject)

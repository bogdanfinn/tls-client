const ffi = require('ffi-napi');

// load the tls-client shared package for your OS you are currently running your nodejs script (i'm running on mac)
const tlsClientLibrary = ffi.Library('./../dist/tls-client-darwin-amd64.dylib', {
    'request': [ 'string', [ 'string' ] ]
});

// build the payload which is needed for the shared package
/* full payload example
{
    "sessionId": "reusableSessionId",
    "tlsClientIdentifier": "chrome_103",
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
    "requestBody": "", // needs to be a string. so JSON.stringify(yourRequestBody) here
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
    "proxyUrl": "",
    "headers": {},
    "headerOrder": [],
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

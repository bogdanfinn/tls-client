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
*/
const requestPayload = {
    "ja3String": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
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

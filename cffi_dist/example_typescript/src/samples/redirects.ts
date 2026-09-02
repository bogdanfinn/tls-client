import {TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

const payload: TLSClientRequestPayload = {
    tlsClientIdentifier: 'chrome_150',
    followRedirects: false,
    sessionId: 'my-redirect-session',
    proxyUrl: '',
    headers: {'accept': '*/*'},
    headerOrder: ['accept'],
    requestUrl: 'https://httpbin.org/redirect/1',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};

let response = tlsClient.request(payload);
console.log('followRedirects=false:', response.status, response.target);

// followRedirects can be changed within an existing session.
payload.followRedirects = true;

response = tlsClient.request(payload);
console.log('followRedirects=true:', response.status, response.target);

tlsClient.destroySession({sessionId: 'my-redirect-session'});

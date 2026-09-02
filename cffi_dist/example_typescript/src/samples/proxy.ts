import {TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

const payload: TLSClientRequestPayload = {
    tlsClientIdentifier: 'chrome_150',
    followRedirects: false,
    sessionId: 'my-proxy-session',
    // http://, socks5:// and socks5h:// proxy URLs are all supported.
    proxyUrl: 'http://user:pass@proxy-one.example.com:8000',
    headers: {'accept': '*/*'},
    headerOrder: ['accept'],
    requestUrl: 'https://tls.peet.ws/api/all',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};

let response = tlsClient.request(payload);
console.log('via proxy one:', response.status);

// Swap the proxy for the same session - no new client/session needs to be built.
payload.proxyUrl = 'http://user:pass@proxy-two.example.com:8000';

response = tlsClient.request(payload);
console.log('via proxy two:', response.status);

tlsClient.destroySession({sessionId: 'my-proxy-session'});

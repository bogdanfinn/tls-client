import {TLSClientFetchCookiesForSessionRequestPayload, TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

const sessionId = 'my-cookie-session';

// The server sets a cookie on this request; the session's cookie jar stores
// it automatically and replays it on every later request in this session.
const payload: TLSClientRequestPayload = {
    tlsClientIdentifier: 'chrome_150',
    followRedirects: true,
    sessionId: sessionId,
    proxyUrl: '',
    headers: {'accept': '*/*'},
    headerOrder: ['accept'],
    requestUrl: 'https://httpbin.org/cookies/set?session=abc123',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};

tlsClient.request(payload);

const fetchCookiesPayload: TLSClientFetchCookiesForSessionRequestPayload = {
    sessionId: sessionId,
    url: 'https://httpbin.org',
};

const cookiesInSession = tlsClient.getCookiesFromSession(fetchCookiesPayload);

console.log(cookiesInSession);

tlsClient.destroySession({sessionId: sessionId});

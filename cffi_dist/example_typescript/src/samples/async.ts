import {
    TLSClientFetchCookiesForSessionRequestPayload,
    TLSClientRequestPayload,
    TLSClientResponseData
} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

const payload: TLSClientRequestPayload = {
    tlsClientIdentifier: 'chrome_103',
    followRedirects: false,
    insecureSkipVerify: false,
    withoutCookieJar: false,
    withRandomTLSExtensionOrder: false,
    timeoutSeconds: 30,
    sessionId: 'my-session-id',
    proxyUrl: '',
    headers: {
        'accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9',
        'user-agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36',
        'accept-encoding': 'gzip, deflate, br',
        'accept-language': 'de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7'
    },
    headerOrder: [
        'accept',
        'user-agent',
        'accept-encoding',
        'accept-language'
    ],
    requestUrl: 'https://www.google.com',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};


tlsClient.requestAsync(payload).then((response) => console.log(response))

const fetchCookiesPayload: TLSClientFetchCookiesForSessionRequestPayload = {
    sessionId: 'my-session-id',
    url: 'https://www.google.com',
};

tlsClient.getCookiesFromSessionAsync(fetchCookiesPayload).then((response) => console.log(response));

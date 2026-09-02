import {TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

const payload: TLSClientRequestPayload = {
    tlsClientIdentifier: 'chrome_150',
    followRedirects: false,
    insecureSkipVerify: false,
    withoutCookieJar: false,
    withRandomTLSExtensionOrder: false,
    timeoutSeconds: 30,
    sessionId: 'my-post-session',
    proxyUrl: '',
    headers: {
        'accept': 'application/json',
        'content-type': 'application/json',
        'user-agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36',
    },
    headerOrder: ['accept', 'content-type', 'user-agent'],
    requestUrl: 'https://httpbin.org/post',
    requestMethod: 'POST',
    requestBody: JSON.stringify({foo: 'bar'}),
    requestCookies: []
};

const response = tlsClient.request(payload);

console.log(response.status, response.body);

tlsClient.destroySession({sessionId: 'my-post-session'});

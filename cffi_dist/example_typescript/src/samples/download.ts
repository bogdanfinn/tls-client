import {tmpdir} from 'os';
import {join} from 'path';
import {writeFileSync} from 'fs';
import {TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

const payload: TLSClientRequestPayload = {
    followRedirects: true,
    proxyUrl: '',
    // isByteResponse base64-encodes the response body, which is required
    // for binary payloads like images.
    isByteResponse: true,
    headers: {'accept': '*/*'},
    headerOrder: ['accept'],
    requestUrl: 'https://avatars.githubusercontent.com/u/17678241?v=4',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};

const response = tlsClient.request(payload);

// The body is a data URI (e.g. "data:image/png;base64,...."); drop
// everything up to and including the first comma before decoding.
const base64Data = response.body.substring(response.body.indexOf(',') + 1);

const dest = join(tmpdir(), 'tls-client-example.jpg');
writeFileSync(dest, Buffer.from(base64Data, 'base64'));

console.log(`status: ${response.status}, wrote file to: ${dest}`);

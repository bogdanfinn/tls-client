import {TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

// customTlsClient lets you build a fingerprint from a raw JA3 string instead
// of picking a bundled tlsClientIdentifier.
const payload: TLSClientRequestPayload = {
    followRedirects: false,
    proxyUrl: '',
    customTlsClient: {
        ja3String: '771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-10-11-13-16-23-43-51-65281-45-21,29-23-24,0',
        h2Settings: {
            HEADER_TABLE_SIZE: 65536,
            MAX_CONCURRENT_STREAMS: 1000,
            INITIAL_WINDOW_SIZE: 6291456,
            MAX_HEADER_LIST_SIZE: 262144
        },
        h2SettingsOrder: ['HEADER_TABLE_SIZE', 'MAX_CONCURRENT_STREAMS', 'INITIAL_WINDOW_SIZE', 'MAX_HEADER_LIST_SIZE'],
        supportedSignatureAlgorithms: ['ECDSAWithP256AndSHA256', 'PSSWithSHA256', 'PKCS1WithSHA256'],
        supportedVersions: ['GREASE', '1.3', '1.2'],
        keyShareCurves: ['GREASE', 'X25519'],
        certCompressionAlgos: ['brotli'],
        alpnProtocols: ['h2', 'http/1.1'],
        alpsProtocols: ['h2'],
        pseudoHeaderOrder: [':method', ':authority', ':scheme', ':path'],
        connectionFlow: 15663105,
        priorityFrames: [],
    },
    headers: {'accept': '*/*'},
    headerOrder: ['accept'],
    requestUrl: 'https://tls.peet.ws/api/all',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};

const response = tlsClient.request(payload);

console.log(response.status, response.body);

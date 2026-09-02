import {TLSClientRequestPayload} from "@project/types";
import {TLSClient} from "@project/client";

const tlsClient = new TLSClient()

// Pins are the base64 encoded SHA-256 hashes of a host's public keys.
// Generate them with: hpkp-pins -server=bstn.com:443
const payload: TLSClientRequestPayload = {
    tlsClientIdentifier: 'chrome_150',
    followRedirects: false,
    proxyUrl: '',
    certificatePinningHosts: {
        'bstn.com': [
            'NQvy9sFS99nBqk/nZCUF44hFhshrkvxqYtfrZq3i+Ww=',
            '4a6cPehI7OG6cuDZka5NDZ7FR8a60d3auda+sKfg4Ng=',
            'x4QzPSC810K5/cMjb05Qm4k3Bw5zBn4lTdO/nEW/Td4=',
        ]
    },
    headers: {'accept': '*/*'},
    headerOrder: ['accept'],
    requestUrl: 'https://bstn.com',
    requestMethod: 'GET',
    requestBody: '',
    requestCookies: []
};

const response = tlsClient.request(payload);

// If a pin does not match, "status" is 0 and "body" holds the pinning error
// instead of a remote response.
console.log(response);

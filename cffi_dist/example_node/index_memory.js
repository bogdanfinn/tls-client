const ffi = require("ffi-napi");

class TLS {
    constructor() {
        this.tls = this.initTLS();
    }

    initTLS() {
        let tlsLib = "./../dist/tls-client-darwin-amd64-1.0.0.dylib";
        return ffi.Library(tlsLib, {
            request: ["string", ["string"]],
            freeMemory: ["void", ['string']],
            getCookiesFromSession: ["string", ["string"]],
            destroyAll: ["string", []],
            destroySession: ["string", ["string"]],
        });
    }

    async request(payload, clientType = "chrome_105", followRedirects = true) {
        return new Promise((resolve, reject) => {
            const defaultPayload = {
                tlsClientIdentifier: clientType,
                followRedirects,
                insecureSkipVerify: false,
                withoutCookieJar: false,
                timeoutSeconds: 30,
                ...payload,
            };
            this.tls.request.async(JSON.stringify(defaultPayload), (error, resp) => {
                if (error) reject(error);
                const response = JSON.parse(resp);
                this.tls.freeMemory(response.id)
                resolve(response);
            });
        });
    }
}

function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

class TEST {
    constructor(tls) {
        this.tls = tls;
    }

    async init() {
        while (true) {
            await this.monitor();
            await sleep(3000);
        }
    }

    async getPicks() {
        const payload = {
            requestMethod: "GET",
            requestBody: null,
            requestUrl: "https://microsoft.com",
            headerOrder: [
                "authority",
                "sec-ch-ua",
                "ot-tracer-sampled",
                "sec-ch-ua-mobile",
                "tmps-correlation-id",
                "accept",
                "origin",
                "sec-fetch-site",
                "sec-fetch-mode",
                "referer",
                "accept-encoding",
                "accept-language",
                "cookie",
            ],
            headers: {
                "sec-ch-ua": '" Not A;Brand";v="99", "Chromium";v="102", "Google Chrome";v="102"',
                "ot-tracer-sampled": "true",
                "sec-ch-ua-mobile": "?0",
                accept: "*/*",
                origin: "https://tls.peet.ws/api/all",
                "sec-fetch-site": "same-site",
                "sec-fetch-mode": "cors",
                referer: "https://tls.peet.ws/api/all",
                "accept-encoding": "gzip, deflate, br",
                "accept-language": "en-US,en;q=0.9",
            },
        };
        const response = await this.tls.request(payload);
        console.log(response.status);
    }

    async monitor() {
        try {
            console.log(`Monitoring...`);
            await this.getPicks();
        } catch (error) {
            console.log(`[${this.query}]`, error);
        }
    }
}

const tls = new TLS();

const task = new TEST(tls);

task.init();
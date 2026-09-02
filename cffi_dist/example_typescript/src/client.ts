import {join} from 'path';
import {arch, platform} from 'os';
import {Library, LibraryObject} from 'ffi-napi';
import {
    TLSClientAddCookiesToSessionPayload,
    TLSClientAddCookiesToSessionResponse,
    TLSClientDestroyAllResponse,
    TLSClientFetchCookiesForSessionRequestPayload, TLSClientFetchCookiesForSessionResponse,
    TLSClientInstance,
    TLSClientReleaseSessionPayload,
    TLSClientReleaseSessionResponse,
    TLSClientRequestPayload,
    TLSClientResponseData
} from "@project/types";

export class TLSClient implements TLSClientInstance {
    private wrapper: LibraryObject<never> | null

    constructor() {
        this.wrapper = createWrapper()
    }

    request(payload: TLSClientRequestPayload): TLSClientResponseData {
        const resp = this.wrapper.request(JSON.stringify(payload))
        return this.parseAndFree<TLSClientResponseData>(resp)
    }

    async requestAsync(payload: TLSClientRequestPayload): Promise<TLSClientResponseData> {
        return new Promise((resolve) => {
            this.wrapper.request.async(JSON.stringify(payload), (error: Error, response: string) => {
                resolve(this.parseAndFree<TLSClientResponseData>(response));
            })
        })
    }

    destroySession(payload: TLSClientReleaseSessionPayload): TLSClientReleaseSessionResponse {
        const resp = this.wrapper.destroySession(JSON.stringify(payload))
        return this.parseAndFree<TLSClientReleaseSessionResponse>(resp)
    }

    async destroySessionAsync(payload: TLSClientReleaseSessionPayload): Promise<TLSClientReleaseSessionResponse> {
        return new Promise((resolve) => {
            this.wrapper.destroySession.async(JSON.stringify(payload), (error: Error, response: string) => {
                resolve(this.parseAndFree<TLSClientReleaseSessionResponse>(response));
            })
        });
    }

    getCookiesFromSession(payload: TLSClientFetchCookiesForSessionRequestPayload): TLSClientFetchCookiesForSessionResponse {
        const resp = this.wrapper.getCookiesFromSession(JSON.stringify(payload))
        return this.parseAndFree<TLSClientFetchCookiesForSessionResponse>(resp)
    }

    async getCookiesFromSessionAsync(payload: TLSClientFetchCookiesForSessionRequestPayload): Promise<TLSClientFetchCookiesForSessionResponse> {
        return new Promise((resolve) => {
            this.wrapper.getCookiesFromSession.async(JSON.stringify(payload), (error: Error, response: string) => {
                resolve(this.parseAndFree<TLSClientFetchCookiesForSessionResponse>(response));
            })
        });
    }

    addCookiesToSession(payload: TLSClientAddCookiesToSessionPayload): TLSClientAddCookiesToSessionResponse {
        const resp = this.wrapper.addCookiesToSession(JSON.stringify(payload))
        return this.parseAndFree<TLSClientAddCookiesToSessionResponse>(resp)
    }

    async addCookiesToSessionAsync(payload: TLSClientAddCookiesToSessionPayload): Promise<TLSClientAddCookiesToSessionResponse> {
        return new Promise((resolve) => {
            this.wrapper.addCookiesToSession.async(JSON.stringify(payload), (error: Error, response: string) => {
                resolve(this.parseAndFree<TLSClientAddCookiesToSessionResponse>(response));
            })
        });
    }

    // Releases every session the shared library holds. Useful on shutdown, or
    // when you lost track of the session ids you created.
    destroyAll(): TLSClientDestroyAllResponse {
        const resp = this.wrapper.destroyAll()
        return this.parseAndFree<TLSClientDestroyAllResponse>(resp)
    }

    async destroyAllAsync(): Promise<TLSClientDestroyAllResponse> {
        return new Promise((resolve) => {
            this.wrapper.destroyAll.async((error: Error, response: string) => {
                resolve(this.parseAndFree<TLSClientDestroyAllResponse>(response));
            })
        });
    }

    // Every call into the shared library allocates memory on the go side and
    // hands back an "id" for it. That memory is only released when freeMemory
    // is called with the id, so the wrapper does it here as soon as the
    // response has been parsed - otherwise it accumulates for the life of the
    // process.
    private parseAndFree<T extends { id?: string }>(response: string): T {
        const parsed = JSON.parse(response) as T;

        if (parsed.id) {
            this.wrapper.freeMemory(parsed.id);
        }

        return parsed;
    }
}

// Release assets are published as
// tls-client-xgo-<version>-<goos>-<goarch>.<ext>, see
// https://github.com/bogdanfinn/tls-client/releases - bump this when you
// download a newer shared library.
const TLS_CLIENT_VERSION = '1.16.0';

// node names platforms and architectures differently than the go toolchain the
// shared library was built with, so the pairs are mapped explicitly. Anything
// missing here has no published asset rather than a different name.
const SHARED_LIBRARY_TARGETS: { [nodeTarget: string]: string } = {
    'darwin-x64': 'darwin-amd64.dylib',
    'darwin-arm64': 'darwin-arm64.dylib',
    'linux-x64': 'linux-amd64.so',
    'linux-arm64': 'linux-arm64.so',
    'linux-ia32': 'linux-386.so',
    // xgo publishes arm-5, arm-6 and arm-7; arm-7 covers current devices.
    'linux-arm': 'linux-arm-7.so',
    'win32-x64': 'windows-amd64.dll',
    'win32-ia32': 'windows-386.dll'
};

const sharedLibraryFilename = (): string => {
    const nodeTarget = `${platform()}-${arch()}`;
    const target = SHARED_LIBRARY_TARGETS[nodeTarget];

    if (!target) {
        throw new Error(`no tls-client shared library is published for ${nodeTarget}`);
    }

    return `tls-client-xgo-${TLS_CLIENT_VERSION}-${target}`;
};

const createWrapper = (): LibraryObject<never> => {
    const sharedLibraryPath = join(__dirname, './../../dist/');
    const libraryFilename = sharedLibraryFilename();

    return Library(join(sharedLibraryPath, libraryFilename), {
        request: ['string', ['string']],
        getCookiesFromSession: ['string', ['string']],
        addCookiesToSession: ['string', ['string']],
        freeMemory: ["void", ['string']],
        destroyAll: ['string', []],
        destroySession: ['string', ['string']]
    });
}
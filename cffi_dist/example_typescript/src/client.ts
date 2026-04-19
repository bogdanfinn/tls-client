import {join} from 'path';
import {arch, platform} from 'os';
import {Library, LibraryObject} from 'ffi-napi';
import {
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
        return JSON.parse(resp) as TLSClientResponseData
    }

    async requestAsync(payload: TLSClientRequestPayload): Promise<TLSClientResponseData> {
        return new Promise((resolve) => {
            this.wrapper.request.async(JSON.stringify(payload), (error: Error, response: string) => {
                const clientResponse: TLSClientResponseData = JSON.parse(response);

                resolve(clientResponse);
            })
        })
    }

    destroySession(payload: TLSClientReleaseSessionPayload): TLSClientReleaseSessionResponse {
        const resp = this.wrapper.destroySession(JSON.stringify(payload))
        return JSON.parse(resp) as TLSClientReleaseSessionResponse
    }

    async destroySessionAsync(payload: TLSClientReleaseSessionPayload): Promise<TLSClientReleaseSessionResponse> {
        return new Promise((resolve) => {
            this.wrapper.destroySession.async(JSON.stringify(payload), (error: Error, response: string) => {
                const clientResponse: TLSClientReleaseSessionResponse = JSON.parse(response);

                resolve(clientResponse);
            })
        });
    }

    getCookiesFromSession(payload: TLSClientFetchCookiesForSessionRequestPayload): TLSClientFetchCookiesForSessionResponse {
        const resp = this.wrapper.getCookiesFromSession(JSON.stringify(payload))
        return JSON.parse(resp) as TLSClientFetchCookiesForSessionResponse
    }

    async getCookiesFromSessionAsync(payload: TLSClientFetchCookiesForSessionRequestPayload): Promise<TLSClientFetchCookiesForSessionResponse> {
        return new Promise((resolve) => {
            this.wrapper.getCookiesFromSession.async(JSON.stringify(payload), (error: Error, response: string) => {
                const clientResponse: TLSClientFetchCookiesForSessionResponse = JSON.parse(response);

                resolve(clientResponse);
            })
        });
    }
}

const createWrapper = (): LibraryObject<never> => {
    const sharedLibraryPath = join(__dirname, './../../dist/');
    const sharedLibraryFilename = platform() === 'win32'
        ? `tls-client-windows-64-1.7.2.dll`
        : arch() === 'arm64'
            ? `tls-client-darwin-arm64-1.7.2.dylib`
            : `tls-client-darwin-amd64-1.7.2.dylib`;

    return Library(join(sharedLibraryPath, sharedLibraryFilename), {
        request: ['string', ['string']],
        getCookiesFromSession: ['string', ['string']],
        addCookiesToSession: ['string', ['string']],
        freeMemory: ["void", ['string']],
        destroyAll: ['string', []],
        destroySession: ['string', ['string']]
    });
}
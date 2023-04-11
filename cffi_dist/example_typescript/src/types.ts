type TLSClientIdentifier = 'chrome_103';
type TLSClientRequestMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

export interface TLSClientInstance {
    request: (payload: TLSClientRequestPayload) => TLSClientResponseData
    requestAsync: (payload: TLSClientRequestPayload) => Promise<TLSClientResponseData>
    getCookiesFromSession: (payload: TLSClientFetchCookiesForSessionRequestPayload) => TLSClientFetchCookiesForSessionResponse
    getCookiesFromSessionAsync: (payload: TLSClientFetchCookiesForSessionRequestPayload) => Promise<TLSClientFetchCookiesForSessionResponse>
    destroySession: (payload: TLSClientReleaseSessionPayload) => TLSClientReleaseSessionResponse
    destroySessionAsync: (payload: TLSClientReleaseSessionPayload) => Promise<TLSClientReleaseSessionResponse>
}

export interface TLSClientRequestPayload {
    requestUrl: string;
    requestMethod: TLSClientRequestMethod;
    requestBody: string;
    requestCookies?: { [key: string]: string }[]
    tlsClientIdentifier?: TLSClientIdentifier;
    followRedirects?: boolean;
    insecureSkipVerify?: boolean;
    isByteResponse?: boolean;
    withoutCookieJar?: boolean;
    withRandomTLSExtensionOrder?: boolean;
    timeoutSeconds?: number;
    sessionId?: string;
    proxyUrl?: string;
    headers?: { [key: string]: string };
    headerOrder?: string[];
    customTlsClient?: {
        ja3String: string;
        h2Settings: {
            HEADER_TABLE_SIZE: number;
            MAX_CONCURRENT_STREAMS: number;
            INITIAL_WINDOW_SIZE: number;
            MAX_HEADER_LIST_SIZE: number;
        },
        h2SettingsOrder: string[];
        supportedSignatureAlgorithms: string[];
        supportedVersions: string[];
        keyShareCurves: string[];
        certCompressionAlgo: string;
        pseudoHeaderOrder: string[];
        connectionFlow: number;
        priorityFrames: string[]
    }
}

export interface TLSClientResponseData {
    sessionId?: string;
    status: number;
    target: string;
    body: string;
    headers: { [key: string]: string[] };
    cookies: { [key: string]: string };
}

export interface TLSClientReleaseSessionPayload {
    sessionId: string;
}

export type TLSClientReleaseSessionResponse = {
    success: boolean;
};

export interface TLSClientFetchCookiesForSessionRequestPayload {
    sessionId: string;
    url: string;
}

export type TLSClientFetchCookiesForSessionResponse = Cookie[];

export interface Cookie {
    Name: string;
    Value: string;
    Path: string;
    Domain: string;
    Expires: string;
    RawExpires: string;
    MaxAge: number;
    Secure: boolean;
    HttpOnly: boolean;
    SameSite: number;
    Raw: string;
    Unparsed: string;
}

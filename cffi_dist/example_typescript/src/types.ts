// These types mirror the payload and response structs of the shared library,
// which live in `cffi_src/types.go` of the tls-client repository. When you
// upgrade the shared library, check that file for new fields.

// Any identifier from "Supported and tested Client Profiles" in the docs,
// e.g. 'chrome_150', 'firefox_148'. Omit it when you supply a customTlsClient.
export type TLSClientIdentifier = string;

export type TLSClientRequestMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS';

// The HTTP/2 SETTINGS the shared library knows by name; anything else is
// ignored when the client profile is built.
export type H2Setting =
    | 'HEADER_TABLE_SIZE'
    | 'ENABLE_PUSH'
    | 'MAX_CONCURRENT_STREAMS'
    | 'INITIAL_WINDOW_SIZE'
    | 'MAX_FRAME_SIZE'
    | 'MAX_HEADER_LIST_SIZE'
    | 'UNKNOWN_SETTING_7'
    | 'UNKNOWN_SETTING_8'
    | 'UNKNOWN_SETTING_9';

// The HTTP/3 SETTINGS the shared library knows by name.
export type H3Setting =
    | 'QPACK_MAX_TABLE_CAPACITY'
    | 'MAX_FIELD_SECTION_SIZE'
    | 'QPACK_BLOCKED_STREAMS'
    | 'H3_DATAGRAM';

export type PseudoHeader = ':method' | ':authority' | ':scheme' | ':path';

export interface PriorityParam {
    streamDep: number;
    exclusive: boolean;
    weight: number;
}

export interface PriorityFrame {
    streamID: number;
    priorityParam: PriorityParam;
}

export interface CandidateCipherSuite {
    kdfId: string;
    aeadId: string;
}

// Settings for the underlying http transport of the client.
export interface TransportOptions {
    // Nanoseconds, because it maps onto go's time.Duration. Zero means no limit.
    idleConnTimeout?: number;
    maxIdleConns?: number;
    maxIdleConnsPerHost?: number;
    maxConnsPerHost?: number;
    // Zero means to use a default limit.
    maxResponseHeaderBytes?: number;
    // If zero, a default (currently 4KB) is used.
    writeBufferSize?: number;
    // If zero, a default (currently 4KB) is used.
    readBufferSize?: number;
    disableKeepAlives?: boolean;
    disableCompression?: boolean;
}

// A fingerprint built from a raw JA3 string, used instead of a bundled
// tlsClientIdentifier. Only ja3String is required; every other field falls back
// to a zero value on the go side.
export interface CustomTlsClient {
    ja3String: string;
    // Required when the ja3String lists extension 51764 (trust_anchors), as a
    // ja3 string carries extension IDs but no extension data. Take it from the
    // data field of the "Unknown extension 51764" entry of a browser
    // fingerprint; "0000" is an empty anchor list.
    trustAnchorsPayload?: string;
    supportedSignatureAlgorithms?: string[];
    supportedDelegatedCredentialsAlgorithms?: string[];
    supportedVersions?: string[];
    keyShareCurves?: string[];
    certCompressionAlgos?: string[];
    alpnProtocols?: string[];
    alpsProtocols?: string[];
    recordSizeLimit?: number;
    ECHCandidatePayloads?: number[];
    ECHCandidateCipherSuites?: CandidateCipherSuite[];

    // HTTP/2
    h2Settings?: Partial<Record<H2Setting, number>>;
    h2SettingsOrder?: H2Setting[];
    pseudoHeaderOrder?: PseudoHeader[];
    connectionFlow?: number;
    priorityFrames?: PriorityFrame[];
    headerPriority?: PriorityParam;
    streamId?: number;

    // HTTP/3. Leaving these out means the profile has no HTTP/3 fingerprint,
    // which only matters when the client actually uses HTTP/3.
    h3Settings?: Partial<Record<H3Setting, number>>;
    h3SettingsOrder?: H3Setting[];
    h3PseudoHeaderOrder?: PseudoHeader[];
    h3PriorityParam?: number;
    h3SendGreaseFrames?: boolean;

    allowHttp?: boolean;
}

// The cookie shape the shared library accepts on a request and returns from
// getCookiesFromSession.
export interface Cookie {
    name: string;
    value: string;
    path?: string;
    domain?: string;
    // Unix timestamp in seconds, not an ISO string.
    expires?: number;
    maxAge?: number;
    secure?: boolean;
    httpOnly?: boolean;
}

export interface TLSClientRequestPayload {
    requestUrl: string;
    requestMethod: TLSClientRequestMethod;
    requestBody?: string;
    requestCookies?: Cookie[];
    // Sends requestBody as base64 encoded bytes instead of a string.
    isByteRequest?: boolean;
    // Returns the response body as base64 encoded bytes instead of a string.
    isByteResponse?: boolean;
    // Overrides the Host header without changing the requestUrl.
    requestHostOverride?: string;

    tlsClientIdentifier?: TLSClientIdentifier;
    customTlsClient?: CustomTlsClient;
    withRandomTLSExtensionOrder?: boolean;
    transportOptions?: TransportOptions;

    // Protocol selection. withProtocolRacing races HTTP/3 and HTTP/2 and
    // requires a socks5:// proxy when proxyUrl is set, because only SOCKS5 can
    // carry the UDP traffic HTTP/3 needs.
    forceHttp1?: boolean;
    disableHttp3?: boolean;
    withProtocolRacing?: boolean;
    disableSessionTickets?: boolean;

    followRedirects?: boolean;
    insecureSkipVerify?: boolean;
    certificatePinningHosts?: { [host: string]: string[] };
    serverNameOverwrite?: string;

    proxyUrl?: string;
    isRotatingProxy?: boolean;
    localAddress?: string;
    disableIPV4?: boolean;
    disableIPV6?: boolean;

    // Single valued, unlike defaultHeaders and connectHeaders.
    headers?: { [key: string]: string };
    headerOrder?: string[];
    defaultHeaders?: { [key: string]: string[] };
    // Headers sent on the CONNECT request to an HTTP proxy.
    connectHeaders?: { [key: string]: string[] };

    timeoutSeconds?: number;
    timeoutMilliseconds?: number;

    sessionId?: string;
    withoutCookieJar?: boolean;
    withCustomCookieJar?: boolean;

    // Streams the response body to streamOutputPath instead of returning it in
    // the response, writing streamOutputBlockSize bytes at a time and
    // terminating with streamOutputEOFSymbol.
    streamOutputPath?: string;
    streamOutputBlockSize?: number;
    streamOutputEOFSymbol?: string;

    withDebug?: boolean;
    catchPanics?: boolean;
}

export interface TLSClientResponseData {
    // The handle for the memory this response allocated on the go side. Pass
    // it to freeMemory once you are done with the response.
    id: string;
    sessionId?: string;
    status: number;
    target: string;
    body: string;
    // The protocol the request was actually sent over, e.g. "HTTP/2.0".
    usedProtocol: string;
    headers: { [key: string]: string[] };
    cookies: { [key: string]: string };
}

export interface TLSClientReleaseSessionPayload {
    sessionId: string;
}

export type TLSClientReleaseSessionResponse = {
    id: string;
    success: boolean;
};

export interface TLSClientFetchCookiesForSessionRequestPayload {
    sessionId: string;
    url: string;
}

export type TLSClientFetchCookiesForSessionResponse = { id: string, cookies: Cookie[] };

export interface TLSClientAddCookiesToSessionPayload {
    sessionId: string;
    url: string;
    cookies: Cookie[];
}

// addCookiesToSession answers with the full cookie list of the session, the
// same shape getCookiesFromSession returns.
export type TLSClientAddCookiesToSessionResponse = TLSClientFetchCookiesForSessionResponse;

export type TLSClientDestroyAllResponse = {
    id: string;
    success: boolean;
};

export interface TLSClientInstance {
    request: (payload: TLSClientRequestPayload) => TLSClientResponseData
    requestAsync: (payload: TLSClientRequestPayload) => Promise<TLSClientResponseData>
    getCookiesFromSession: (payload: TLSClientFetchCookiesForSessionRequestPayload) => TLSClientFetchCookiesForSessionResponse
    getCookiesFromSessionAsync: (payload: TLSClientFetchCookiesForSessionRequestPayload) => Promise<TLSClientFetchCookiesForSessionResponse>
    addCookiesToSession: (payload: TLSClientAddCookiesToSessionPayload) => TLSClientAddCookiesToSessionResponse
    addCookiesToSessionAsync: (payload: TLSClientAddCookiesToSessionPayload) => Promise<TLSClientAddCookiesToSessionResponse>
    destroySession: (payload: TLSClientReleaseSessionPayload) => TLSClientReleaseSessionResponse
    destroySessionAsync: (payload: TLSClientReleaseSessionPayload) => Promise<TLSClientReleaseSessionResponse>
    destroyAll: () => TLSClientDestroyAllResponse
    destroyAllAsync: () => Promise<TLSClientDestroyAllResponse>
}

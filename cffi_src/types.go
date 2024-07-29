package tls_client_cffi_src

import (
	"encoding/json"
	"fmt"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
)

type TLSClientError struct {
	err error
}

func NewTLSClientError(err error) *TLSClientError {
	return &TLSClientError{
		err: err,
	}
}

func (e *TLSClientError) Error() string {
	return e.err.Error()
}

type DestroySessionInput struct {
	SessionId string `json:"sessionId"`
}

type DestroyOutput struct {
	Id      string `json:"id"`
	Success bool   `json:"success"`
}

type AddCookiesToSessionInput struct {
	Cookies   []Cookie `json:"cookies"`
	SessionId string   `json:"sessionId"`
	Url       string   `json:"url"`
}

type GetCookiesFromSessionInput struct {
	SessionId string `json:"sessionId"`
	Url       string `json:"url"`
}

type CookiesFromSessionOutput struct {
	Id      string   `json:"id"`
	Cookies []Cookie `json:"cookies"`
}

// RequestInput is the data a Python client can construct a client and request from.
type RequestInput struct {
	CatchPanics                 bool                `json:"catchPanics"`
	CertificatePinningHosts     map[string][]string `json:"certificatePinningHosts"`
	CustomTlsClient             *CustomTlsClient    `json:"customTlsClient"`
	TransportOptions            *TransportOptions   `json:"transportOptions"`
	FollowRedirects             bool                `json:"followRedirects"`
	ForceHttp1                  bool                `json:"forceHttp1"`
	HeaderOrder                 []string            `json:"headerOrder"`
	Headers                     map[string]string   `json:"headers"`
	DefaultHeaders              map[string][]string `json:"defaultHeaders"`
	ConnectHeaders              map[string][]string `json:"connectHeaders"`
	InsecureSkipVerify          bool                `json:"insecureSkipVerify"`
	IsByteRequest               bool                `json:"isByteRequest"`
	IsByteResponse              bool                `json:"isByteResponse"`
	IsRotatingProxy             bool                `json:"isRotatingProxy"`
	DisableIPV6                 bool                `json:"disableIPV6"`
	DisableIPV4                 bool                `json:"disableIPV4"`
	LocalAddress                *string             `json:"localAddress"`
	ServerNameOverwrite         *string             `json:"serverNameOverwrite"`
	ProxyUrl                    *string             `json:"proxyUrl"`
	RequestBody                 *string             `json:"requestBody"`
	RequestCookies              []Cookie            `json:"requestCookies"`
	RequestMethod               string              `json:"requestMethod"`
	RequestUrl                  string              `json:"requestUrl"`
	RequestHostOverride         *string             `json:"requestHostOverride"`
	SessionId                   *string             `json:"sessionId"`
	StreamOutputBlockSize       *int                `json:"streamOutputBlockSize"`
	StreamOutputEOFSymbol       *string             `json:"streamOutputEOFSymbol"`
	StreamOutputPath            *string             `json:"streamOutputPath"`
	TimeoutMilliseconds         int                 `json:"timeoutMilliseconds"`
	TimeoutSeconds              int                 `json:"timeoutSeconds"`
	TLSClientIdentifier         string              `json:"tlsClientIdentifier"`
	WithDebug                   bool                `json:"withDebug"`
	WithDefaultCookieJar        bool                `json:"withDefaultCookieJar"`
	WithoutCookieJar            bool                `json:"withoutCookieJar"`
	WithRandomTLSExtensionOrder bool                `json:"withRandomTLSExtensionOrder"`
}

// CustomTlsClient contains custom TLS specifications to construct a client from.
type CustomTlsClient struct {
	CertCompressionAlgo                     string                `json:"certCompressionAlgo"`
	ConnectionFlow                          uint32                `json:"connectionFlow"`
	H2Settings                              map[string]uint32     `json:"h2Settings"`
	H2SettingsOrder                         []string              `json:"h2SettingsOrder"`
	HeaderPriority                          *PriorityParam        `json:"headerPriority"`
	Ja3String                               string                `json:"ja3String"`
	KeyShareCurves                          []string              `json:"keyShareCurves"`
	ALPNProtocols                           []string              `json:"alpnProtocols"`
	ALPSProtocols                           []string              `json:"alpsProtocols"`
	ECHCandidatePayloads                    []uint16              `json:"ECHCandidatePayloads"`
	ECHCandidateCipherSuites                CandidateCipherSuites `json:"ECHCandidateCipherSuites"`
	PriorityFrames                          []PriorityFrames      `json:"priorityFrames"`
	PseudoHeaderOrder                       []string              `json:"pseudoHeaderOrder"`
	SupportedDelegatedCredentialsAlgorithms []string              `json:"supportedDelegatedCredentialsAlgorithms"`
	SupportedSignatureAlgorithms            []string              `json:"supportedSignatureAlgorithms"`
	SupportedVersions                       []string              `json:"supportedVersions"`
}

type CandidateCipherSuites []CandidateCipherSuite

func (c CandidateCipherSuites) Translate() []tls_client.CandidateCipherSuites {
	suites := make([]tls_client.CandidateCipherSuites, len(c))
	for i, suite := range c {
		suites[i] = tls_client.CandidateCipherSuites{
			KdfId:  suite.KdfId,
			AeadId: suite.AeadId,
		}
	}

	return suites
}

type CandidateCipherSuite struct {
	KdfId  string `json:"kdfId"`
	AeadId string `json:"aeadId"`
}

// TransportOptions contains settings for the underlying http transport of the tls client
type TransportOptions struct {
	DisableKeepAlives      bool  `json:"disableKeepAlives"`
	DisableCompression     bool  `json:"disableCompression"`
	MaxIdleConns           int   `json:"maxIdleConns"`
	MaxIdleConnsPerHost    int   `json:"maxIdleConnsPerHost"`
	MaxConnsPerHost        int   `json:"maxConnsPerHost"`
	MaxResponseHeaderBytes int64 `json:"maxResponseHeaderBytes"` // Zero means to use a default limit.
	WriteBufferSize        int   `json:"writeBufferSize"`        // If zero, a default (currently 4KB) is used.
	ReadBufferSize         int   `json:"readBufferSize"`         // If zero, a default (currently 4KB) is used.
	// IdleConnTimeout is the maximum amount of time an idle (keep-alive)
	// connection will remain idle before closing itself. Zero means no limit.
	IdleConnTimeout *time.Duration `json:"idleConnTimeout"`
}

type PriorityFrames struct {
	PriorityParam PriorityParam `json:"priorityParam"`
	StreamID      uint32        `json:"streamID"`
}

type PriorityParam struct {
	Exclusive bool   `json:"exclusive"`
	StreamDep uint32 `json:"streamDep"`
	Weight    uint8  `json:"weight"`
}

type Cookie struct {
	Domain  string    `json:"domain"`
	Expires Timestamp `json:"expires"`
	MaxAge  int       `json:"maxAge"`
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Value   string    `json:"value"`
}

type Timestamp struct {
	time.Time
}

func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
	var raw int64
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return fmt.Errorf("error decoding timestamp: %w", err)
	}

	p.Time = time.Unix(raw, 0)

	return nil
}

func (p *Timestamp) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%d", p.Unix())
	return []byte(stamp), nil
}

// Response is the response that is sent back to the Python client.
type Response struct {
	Id           string              `json:"id"`
	Body         string              `json:"body"`
	Cookies      map[string]string   `json:"cookies"`
	Headers      map[string][]string `json:"headers"`
	SessionId    string              `json:"sessionId,omitempty"`
	Status       int                 `json:"status"`
	Target       string              `json:"target"`
	UsedProtocol string              `json:"usedProtocol"`
}

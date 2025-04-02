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
	SessionId string   `json:"sessionId"`
	Url       string   `json:"url"`
	Cookies   []Cookie `json:"cookies"`
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
	CertificatePinningHosts     map[string][]string `json:"certificatePinningHosts"`
	CustomTlsClient             *CustomTlsClient    `json:"customTlsClient"`
	TransportOptions            *TransportOptions   `json:"transportOptions"`
	Headers                     map[string]string   `json:"headers"`
	DefaultHeaders              map[string][]string `json:"defaultHeaders"`
	ConnectHeaders              map[string][]string `json:"connectHeaders"`
	LocalAddress                *string             `json:"localAddress"`
	ServerNameOverwrite         *string             `json:"serverNameOverwrite"`
	ProxyUrl                    *string             `json:"proxyUrl"`
	RequestBody                 *string             `json:"requestBody"`
	RequestHostOverride         *string             `json:"requestHostOverride"`
	SessionId                   *string             `json:"sessionId"`
	StreamOutputBlockSize       *int                `json:"streamOutputBlockSize"`
	StreamOutputEOFSymbol       *string             `json:"streamOutputEOFSymbol"`
	StreamOutputPath            *string             `json:"streamOutputPath"`
	RequestMethod               string              `json:"requestMethod"`
	RequestUrl                  string              `json:"requestUrl"`
	TLSClientIdentifier         string              `json:"tlsClientIdentifier"`
	HeaderOrder                 []string            `json:"headerOrder"`
	RequestCookies              []Cookie            `json:"requestCookies"`
	TimeoutMilliseconds         int                 `json:"timeoutMilliseconds"`
	TimeoutSeconds              int                 `json:"timeoutSeconds"`
	CatchPanics                 bool                `json:"catchPanics"`
	FollowRedirects             bool                `json:"followRedirects"`
	ForceHttp1                  bool                `json:"forceHttp1"`
	InsecureSkipVerify          bool                `json:"insecureSkipVerify"`
	IsByteRequest               bool                `json:"isByteRequest"`
	IsByteResponse              bool                `json:"isByteResponse"`
	IsRotatingProxy             bool                `json:"isRotatingProxy"`
	DisableIPV6                 bool                `json:"disableIPV6"`
	DisableIPV4                 bool                `json:"disableIPV4"`
	WithDebug                   bool                `json:"withDebug"`
	WithDefaultCookieJar        bool                `json:"withDefaultCookieJar"`
	WithoutCookieJar            bool                `json:"withoutCookieJar"`
	WithRandomTLSExtensionOrder bool                `json:"withRandomTLSExtensionOrder"`
}

// CustomTlsClient contains custom TLS specifications to construct a client from.
type CustomTlsClient struct {
	H2Settings                              map[string]uint32     `json:"h2Settings"`
	HeaderPriority                          *PriorityParam        `json:"headerPriority"`
	CertCompressionAlgos                    []string              `json:"certCompressionAlgos"`
	Ja3String                               string                `json:"ja3String"`
	H2SettingsOrder                         []string              `json:"h2SettingsOrder"`
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
	ConnectionFlow                          uint32                `json:"connectionFlow"`
	RecordSizeLimit                         uint16                `json:"recordSizeLimit"`
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
	// IdleConnTimeout is the maximum amount of time an idle (keep-alive)
	// connection will remain idle before closing itself. Zero means no limit.
	IdleConnTimeout        *time.Duration `json:"idleConnTimeout"`
	MaxIdleConns           int            `json:"maxIdleConns"`
	MaxIdleConnsPerHost    int            `json:"maxIdleConnsPerHost"`
	MaxConnsPerHost        int            `json:"maxConnsPerHost"`
	MaxResponseHeaderBytes int64          `json:"maxResponseHeaderBytes"` // Zero means to use a default limit.
	WriteBufferSize        int            `json:"writeBufferSize"`        // If zero, a default (currently 4KB) is used.
	ReadBufferSize         int            `json:"readBufferSize"`         // If zero, a default (currently 4KB) is used.
	DisableKeepAlives      bool           `json:"disableKeepAlives"`
	DisableCompression     bool           `json:"disableCompression"`
}

type PriorityFrames struct {
	PriorityParam PriorityParam `json:"priorityParam"`
	StreamID      uint32        `json:"streamID"`
}

type PriorityParam struct {
	StreamDep uint32 `json:"streamDep"`
	Exclusive bool   `json:"exclusive"`
	Weight    uint8  `json:"weight"`
}

type Cookie struct {
	Expires  Timestamp `json:"expires"`
	Domain   string    `json:"domain"`
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Value    string    `json:"value"`
	MaxAge   int       `json:"maxAge"`
	Secure   bool      `json:"secure"`
	HttpOnly bool      `json:"httpOnly"`
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
	Cookies      map[string]string   `json:"cookies"`
	Headers      map[string][]string `json:"headers"`
	Id           string              `json:"id"`
	Body         string              `json:"body"`
	SessionId    string              `json:"sessionId,omitempty"`
	Target       string              `json:"target"`
	UsedProtocol string              `json:"usedProtocol"`
	Status       int                 `json:"status"`
}

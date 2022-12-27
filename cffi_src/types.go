package tls_client_cffi_src

import (
	"encoding/json"
	"fmt"
	"time"

	http "github.com/bogdanfinn/fhttp"
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
	SessionId string         `json:"sessionId"`
	Url       string         `json:"url"`
	Cookies   []*http.Cookie `json:"cookies"`
}

type GetCookiesFromSessionInput struct {
	SessionId string `json:"sessionId"`
	Url       string `json:"url"`
}

type CookiesFromSessionOutput struct {
	Id      string         `json:"id"`
	Cookies []*http.Cookie `json:"cookies"`
}

// RequestInput is the data a Python client can construct a client and request from.
type RequestInput struct {
	SessionId                   *string           `json:"sessionId"`
	TLSClientIdentifier         string            `json:"tlsClientIdentifier"`
	CustomTlsClient             *CustomTlsClient  `json:"customTlsClient"`
	FollowRedirects             bool              `json:"followRedirects"`
	ForceHttp1                  bool              `json:"forceHttp1"`
	IsByteResponse              bool              `json:"isByteResponse"`
	WithDebug                   bool              `json:"withDebug"`
	IsByteRequest               bool              `json:"isByteRequest"`
	WithoutCookieJar            bool              `json:"withoutCookieJar"`
	WithRandomTLSExtensionOrder bool              `json:"withRandomTLSExtensionOrder"`
	InsecureSkipVerify          bool              `json:"insecureSkipVerify"`
	TimeoutSeconds              int               `json:"timeoutSeconds"`
	ProxyUrl                    *string           `json:"proxyUrl"`
	Headers                     map[string]string `json:"headers"`
	HeaderOrder                 []string          `json:"headerOrder"`
	RequestUrl                  string            `json:"requestUrl"`
	RequestMethod               string            `json:"requestMethod"`
	RequestBody                 *string           `json:"requestBody"`
	RequestCookies              []CookieInput     `json:"requestCookies"`
}

// CustomTlsClient contains custom TLS specifications to construct a client from.
type CustomTlsClient struct {
	Ja3String                    string            `json:"ja3String"`
	SupportedSignatureAlgorithms []string          `json:"supportedSignatureAlgorithms"`
	SupportedVersions            []string          `json:"supportedVersions"`
	KeyShareCurves               []string          `json:"keyShareCurves"`
	CertCompressionAlgo          string            `json:"certCompressionAlgo"`
	H2Settings                   map[string]uint32 `json:"h2Settings"`
	H2SettingsOrder              []string          `json:"h2SettingsOrder"`
	PseudoHeaderOrder            []string          `json:"pseudoHeaderOrder"`
	ConnectionFlow               uint32            `json:"connectionFlow"`
	PriorityFrames               []PriorityFrames  `json:"priorityFrames"`
	HeaderPriority               *PriorityParam    `json:"headerPriority"`
}

type PriorityFrames struct {
	StreamID      uint32        `json:"streamID"`
	PriorityParam PriorityParam `json:"priorityParam"`
}

type PriorityParam struct {
	StreamDep uint32 `json:"streamDep"`
	Exclusive bool   `json:"exclusive"`
	Weight    uint8  `json:"weight"`
}

type CookieInput struct {
	Name    string    `json:"name"`
	Value   string    `json:"value"`
	Path    string    `json:"path"`
	Domain  string    `json:"domain"`
	Expires Timestamp `json:"expires"`
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

// Response is the response that is sent back to the Python client.
type Response struct {
	Id        string              `json:"id"`
	SessionId string              `json:"sessionId,omitempty"`
	Status    int                 `json:"status"`
	Target    string              `json:"target"`
	Body      string              `json:"body"`
	Headers   map[string][]string `json:"headers"`
	Cookies   map[string]string   `json:"cookies"`
}

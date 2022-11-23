package tls_client_cffi_src

import (
	"encoding/json"
	"fmt"
	"time"
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

type FreeSessionInput struct {
	SessionId string `json:"sessionId"`
}

type FreeOutput struct {
	Success bool `json:"success"`
}

type GetCookiesFromSessionInput struct {
	SessionId string `json:"sessionId"`
	Url       string `json:"url"`
}

type RequestInput struct {
	SessionId                   *string           `json:"sessionId"`
	TLSClientIdentifier         string            `json:"tlsClientIdentifier"`
	CustomTlsClient             *CustomTlsClient  `json:"customTlsClient"`
	FollowRedirects             bool              `json:"followRedirects"`
	IsByteResponse              bool              `json:"isByteResponse"`
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
}

type PriorityFrames struct {
	StreamID      uint32 `json:"streamID"`
	PriorityParam struct {
		StreamDep uint32 `json:"streamDep"`
		Exclusive bool   `json:"exclusive"`
		Weight    uint8  `json:"weight"`
	} `json:"priorityParam"`
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

	*&p.Time = time.Unix(raw, 0)
	return nil
}

type Response struct {
	SessionId string              `json:"sessionId,omitempty"`
	Status    int                 `json:"status"`
	Target    string              `json:"target"`
	Body      string              `json:"body"`
	Headers   map[string][]string `json:"headers"`
	Cookies   map[string]string   `json:"cookies"`
}

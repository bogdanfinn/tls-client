package main

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

type RequestParams struct {
	SessionId           *string           `json:"sessionId"`
	TLSClientIdentifier string            `json:"tlsClientIdentifier"`
	FollowRedirects     bool              `json:"followRedirects"`
	TimeoutSeconds      int               `json:"timeoutSeconds"`
	Ja3String           string            `json:"ja3String"`
	ProxyUrl            *string           `json:"proxyUrl"`
	Headers             map[string]string `json:"headers"`
	HeaderOrder         []string          `json:"headerOrder"`
	RequestUrl          string            `json:"requestUrl"`
	RequestMethod       string            `json:"requestMethod"`
	RequestBody         *string           `json:"requestBody"`
	RequestCookies      []CookieInput     `json:"requestCookies"`
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
	SessionId string              `json:"sessionId"`
	Status    int                 `json:"status"`
	Body      string              `json:"body"`
	Headers   map[string][]string `json:"headers"`
	Cookies   map[string]string   `json:"cookies"`
}

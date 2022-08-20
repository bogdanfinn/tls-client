package main

import "time"

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
	Expires time.Time `json:"expires"`
}

type Response struct {
	SessionId string              `json:"sessionId"`
	Status    int                 `json:"status"`
	Body      string              `json:"body"`
	Headers   map[string][]string `json:"headers"`
	Cookies   map[string]string   `json:"cookies"`
}

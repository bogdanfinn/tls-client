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
	TLSClientIdentifier string            `json:"tlsClientIdentifier"`
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
	StatusCode      int                 `json:"statusCode"`
	ResponseBody    string              `json:"responseBody"`
	ResponseHeaders map[string][]string `json:"responseHeaders"`
	ResponseCookies map[string]string   `json:"responseCookies"`
	SessionCookies  map[string]string   `json:"sessionCookies"`
}

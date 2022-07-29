package main

import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

//export request
func request(requestParams *C.char) *C.char {
	requestParamsJson := C.GoString(requestParams)

	requestInput := RequestParams{}
	err := json.Unmarshal([]byte(requestParamsJson), &requestInput)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	tlsClientProfile := getTlsClientProfile(requestInput.TLSClientIdentifier)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tlsClientProfile),
	}

	proxy := requestInput.ProxyUrl

	if proxy != nil && *proxy != "" {
		options = append(options, tls_client.WithProxyUrl(*proxy))
	}

	tlsClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	req, err := buildRequest(requestInput)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	cookies := buildCookies(requestInput.RequestCookies)

	if len(cookies) > 0 {
		tlsClient.SetCookies(req.URL, cookies)
	}

	resp, err := tlsClient.Do(req)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	sessionCookies := tlsClient.GetCookies(req.URL)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	defer resp.Body.Close()

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	response := Response{
		StatusCode:      resp.StatusCode,
		ResponseBody:    string(respBodyBytes),
		ResponseHeaders: resp.Header,
		ResponseCookies: cookiesToMap(resp.Cookies()),
		SessionCookies:  cookiesToMap(sessionCookies),
	}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(nil, clientErr)
	}

	return C.CString(string(jsonResponse))
}

func getTlsClientProfile(tlsClientIdentifier string) tls_client.ClientProfile {
	tlsClientProfile, ok := tls_client.MappedTLSClients[tlsClientIdentifier]

	if !ok {
		return tls_client.DefaultClientProfile
	}

	return tlsClientProfile
}

func buildRequest(input RequestParams) (*http.Request, error) {
	var tlsReq *http.Request
	var err error

	if input.RequestMethod == "" || input.RequestUrl == "" {
		return nil, fmt.Errorf("no request url or request method provided")
	}

	if input.RequestBody != nil && *input.RequestBody != "" {
		requestBody := bytes.NewBuffer([]byte(*input.RequestBody))
		tlsReq, err = http.NewRequest(input.RequestMethod, input.RequestUrl, requestBody)
	} else {
		tlsReq, err = http.NewRequest(input.RequestMethod, input.RequestUrl, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request object: %w", err)
	}

	headers := http.Header{
		http.HeaderOrderKey: input.HeaderOrder,
	}

	for key, value := range input.Headers {
		headers[key] = []string{value}
	}

	tlsReq.Header = headers

	return tlsReq, nil
}

func handleResponse(response interface{}, err error) *C.char {
	if err != nil {
		return C.CString(err.Error())
	}

	out, jsonErr := json.Marshal(response)

	if jsonErr != nil {
		return C.CString(jsonErr.Error())
	}

	return C.CString(string(out))
}

func cookiesToMap(cookies []*http.Cookie) map[string]string {
	ret := make(map[string]string, 0)

	for _, c := range cookies {
		ret[c.Name] = c.String()
	}

	return ret
}

func buildCookies(cookies []CookieInput) []*http.Cookie {
	var ret []*http.Cookie

	for _, cookie := range cookies {
		ret = append(ret, &http.Cookie{
			Name:    cookie.Name,
			Value:   cookie.Value,
			Path:    cookie.Path,
			Domain:  cookie.Domain,
			Expires: cookie.Expires,
		})
	}

	return ret
}

func main() {

}

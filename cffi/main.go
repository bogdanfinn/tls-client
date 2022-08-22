package main

import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/google/uuid"
)

var clientsLock = sync.Mutex{}
var clients = make(map[string]tls_client.HttpClient)

//export request
func request(requestParams *C.char) *C.char {
	requestParamsJson := C.GoString(requestParams)

	requestInput := RequestParams{}
	err := json.Unmarshal([]byte(requestParamsJson), &requestInput)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse("", clientErr)
	}

	if requestInput.TLSClientIdentifier != "" && requestInput.Ja3String != "" {
		clientErr := NewTLSClientError(fmt.Errorf("can not built client out of client identifier and ja3string. Please provide only one of them"))
		return handleResponse("", clientErr)
	}

	if requestInput.TLSClientIdentifier == "" && requestInput.Ja3String == "" {
		clientErr := NewTLSClientError(fmt.Errorf("can not built client out without client identifier and without ja3string. Please provide only one of them"))
		return handleResponse("", clientErr)
	}

	tlsClient, newSessionId, err := getTlsClient(requestInput)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(newSessionId, clientErr)
	}

	req, err := buildRequest(requestInput)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(newSessionId, clientErr)
	}

	cookies := buildCookies(requestInput.RequestCookies)

	if len(cookies) > 0 {
		tlsClient.SetCookies(req.URL, cookies)
	}

	resp, err := tlsClient.Do(req)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(newSessionId, clientErr)
	}

	sessionCookies := tlsClient.GetCookies(req.URL)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(newSessionId, clientErr)
	}

	defer resp.Body.Close()

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(newSessionId, clientErr)
	}

	response := Response{
		SessionId: newSessionId,
		Status:    resp.StatusCode,
		Body:      string(respBodyBytes),
		Headers:   resp.Header,
		Cookies:   cookiesToMap(sessionCookies),
	}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		clientErr := NewTLSClientError(err)
		return handleResponse(newSessionId, clientErr)
	}

	return C.CString(string(jsonResponse))
}

func getTlsClient(requestInput RequestParams) (tls_client.HttpClient, string, error) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	sessionId := requestInput.SessionId
	tlsClientIdentifier := requestInput.TLSClientIdentifier
	ja3String := requestInput.Ja3String
	proxyUrl := requestInput.ProxyUrl

	newSessionId := uuid.New().String()
	if sessionId != nil && *sessionId != "" {
		newSessionId = *sessionId
	}

	client, ok := clients[newSessionId]

	if ok {
		return client, newSessionId, nil
	}

	var clientProfile tls_client.ClientProfile

	if tlsClientIdentifier != "" {
		clientProfile = getTlsClientProfile(tlsClientIdentifier)
	}

	if ja3String != "" {
		var decodeErr error
		clientProfile, decodeErr = tls_client.GetClientProfileFromJa3String(ja3String)

		if decodeErr != nil {
			return nil, newSessionId, fmt.Errorf("can not build http client out of ja3 string: %w", decodeErr)
		}
	}

	timeoutSeconds := 30

	if requestInput.TimeoutSeconds != 0 {
		timeoutSeconds = requestInput.TimeoutSeconds
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(timeoutSeconds),
		tls_client.WithClientProfile(clientProfile),
	}

	if !requestInput.FollowRedirects {
		options = append(options, tls_client.WithNotFollowRedirects())
	}

	proxy := proxyUrl

	if proxy != nil && *proxy != "" {
		options = append(options, tls_client.WithProxyUrl(*proxy))
	}

	tlsClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	clients[newSessionId] = tlsClient

	return tlsClient, newSessionId, err
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

func handleResponse(sessionId string, err error) *C.char {
	response := Response{
		SessionId: sessionId,
		Status:    0,
		Body:      err.Error(),
		Headers:   nil,
		Cookies:   nil,
	}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		return C.CString(err.Error())
	}

	return C.CString(string(jsonResponse))
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

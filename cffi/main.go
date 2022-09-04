package main

import "C"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	tls "github.com/bogdanfinn/utls"
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

	if requestInput.TLSClientIdentifier != "" && requestInput.CustomTlsClient != nil {
		clientErr := NewTLSClientError(fmt.Errorf("can not built client out of client identifier and custom tls client information. Please provide only one of them"))
		return handleResponse("", clientErr)
	}

	if requestInput.TLSClientIdentifier == "" && requestInput.CustomTlsClient == nil {
		clientErr := NewTLSClientError(fmt.Errorf("can not built client without client identifier and without custom tls client information. Please provide at least one of them"))
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
	proxyUrl := requestInput.ProxyUrl

	newSessionId := uuid.New().String()
	if sessionId != nil && *sessionId != "" {
		newSessionId = *sessionId
	}

	client, ok := clients[newSessionId]

	if ok {
		modifiedClient, changed, err := handleModification(client, proxyUrl, requestInput.FollowRedirects)
		if err != nil {
			return nil, newSessionId, fmt.Errorf("failed to modify existing client: %w", err)
		}

		if changed {
			clients[newSessionId] = modifiedClient
		}

		return modifiedClient, newSessionId, nil
	}

	var clientProfile tls_client.ClientProfile

	if tlsClientIdentifier != "" {
		clientProfile = getTlsClientProfile(tlsClientIdentifier)
	}

	if requestInput.CustomTlsClient != nil {
		clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, err := getCustomTlsClientProfile(requestInput.CustomTlsClient)

		if err != nil {
			return nil, newSessionId, fmt.Errorf("can not build http client out of custom tls client information: %w", err)
		}

		clientProfile = tls_client.NewClientProfile(clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames)
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

func getCustomTlsClientProfile(customClientDefinition *CustomTlsClient) (tls.ClientHelloID, map[http2.SettingID]uint32, []http2.SettingID, []string, uint32, []http2.Priority, error) {
	specFactory, err := tls_client.GetSpecFactorFromJa3String(customClientDefinition.Ja3String)

	if err != nil {
		return tls.ClientHelloID{}, nil, nil, nil, 0, nil, err
	}

	h2Settings := make(map[http2.SettingID]uint32)
	for key, value := range customClientDefinition.H2Settings {
		h2Settings[http2.SettingID(key)] = value
	}

	var h2SettingsOrder []http2.SettingID
	for _, order := range customClientDefinition.H2SettingsOrder {
		h2SettingsOrder = append(h2SettingsOrder, http2.SettingID(order))
	}

	pseudoHeaderOrder := customClientDefinition.PseudoHeaderOrder
	connectionFlow := customClientDefinition.ConnectionFlow

	var priorityFrames []http2.Priority
	for _, priority := range customClientDefinition.PriorityFrames {
		priorityFrames = append(priorityFrames, http2.Priority{
			StreamID: priority.StreamID,
			PriorityParam: http2.PriorityParam{
				StreamDep: priority.PriorityParam.StreamDep,
				Exclusive: priority.PriorityParam.Exclusive,
				Weight:    priority.PriorityParam.Weight,
			},
		})
	}

	clientHelloId := tls.ClientHelloID{
		Client:      "Custom",
		Version:     "1",
		Seed:        nil,
		SpecFactory: specFactory,
	}

	return clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, nil
}

func getTlsClientProfile(tlsClientIdentifier string) tls_client.ClientProfile {
	tlsClientProfile, ok := tls_client.MappedTLSClients[tlsClientIdentifier]

	if !ok {
		return tls_client.DefaultClientProfile
	}

	return tlsClientProfile
}

func handleModification(client tls_client.HttpClient, proxyUrl *string, followRedirects bool) (tls_client.HttpClient, bool, error) {
	changed := false

	if proxyUrl != nil && client.GetProxy() != *proxyUrl {
		err := client.SetProxy(*proxyUrl)
		if err != nil {
			return nil, false, fmt.Errorf("failed to change proxy url of client: %w", err)
		}

		changed = true
	}

	if client.GetFollowRedirect() != followRedirects {
		client.SetFollowRedirect(followRedirects)
		changed = true
	}

	return client, changed, nil
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
			Expires: cookie.Expires.Time,
		})
	}

	return ret
}

func main() {

}

package tls_client_cffi_src

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"sync"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	tls "github.com/bogdanfinn/utls"
	"github.com/google/uuid"
)

var clientsLock = sync.Mutex{}
var clients = make(map[string]tls_client.HttpClient)

func DestroyTlsClientSession(sessionId string) error {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	_, ok := clients[sessionId]

	if !ok {
		return fmt.Errorf("tls client session with id %s does not exist", sessionId)
	}

	delete(clients, sessionId)

	return nil
}

func DestroyTlsClientSessions() error {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// the remaining clients will be cleaned up by the garbage collection
	clients = make(map[string]tls_client.HttpClient)

	return nil
}

func GetTlsClientFromSession(sessionId string) (tls_client.HttpClient, error) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	client, ok := clients[sessionId]

	if !ok {
		return nil, fmt.Errorf("no client found for sessionId: %s", sessionId)
	}

	return client, nil
}

func GetTlsClientFromInput(requestInput RequestInput) (tls_client.HttpClient, string, bool, *TLSClientError) {
	withSession := true
	sessionId := requestInput.SessionId

	newSessionId := uuid.New().String()
	if sessionId != nil && *sessionId != "" {
		newSessionId = *sessionId
	} else {
		withSession = false
	}

	if requestInput.TLSClientIdentifier != "" && requestInput.CustomTlsClient != nil {
		clientErr := NewTLSClientError(fmt.Errorf("can not built client out of client identifier and custom tls client information. Please provide only one of them"))
		return nil, newSessionId, withSession, clientErr
	}

	if requestInput.TLSClientIdentifier == "" && requestInput.CustomTlsClient == nil {
		clientErr := NewTLSClientError(fmt.Errorf("can not built client without client identifier and without custom tls client information. Please provide at least one of them"))
		return nil, newSessionId, withSession, clientErr
	}

	tlsClient, err := getTlsClient(requestInput, newSessionId, withSession)

	if err != nil {
		clientErr := NewTLSClientError(fmt.Errorf("failed to build client out of request input: %w", err))
		return nil, newSessionId, withSession, clientErr
	}

	return tlsClient, newSessionId, withSession, nil
}

func BuildRequest(input RequestInput) (*http.Request, *TLSClientError) {
	var tlsReq *http.Request
	var err error

	if input.RequestMethod == "" || input.RequestUrl == "" {
		return nil, NewTLSClientError(fmt.Errorf("no request url or request method provided"))
	}

	if input.RequestBody != nil && *input.RequestBody != "" {
		_, ok1 := input.Headers["content-type"]
		_, ok2 := input.Headers["Content-Type"]

		if !ok1 && !ok2 {
			return nil, NewTLSClientError(fmt.Errorf("if you are using a request post body please specify a Content-Type Header"))
		}

		requestBody := bytes.NewBuffer([]byte(*input.RequestBody))
		tlsReq, err = http.NewRequest(input.RequestMethod, input.RequestUrl, requestBody)
	} else {
		tlsReq, err = http.NewRequest(input.RequestMethod, input.RequestUrl, nil)
	}

	if err != nil {
		return nil, NewTLSClientError(fmt.Errorf("failed to create request object: %w", err))
	}

	headers := http.Header{}

	for key, value := range input.Headers {
		headers[key] = []string{value}
	}

	headers[http.HeaderOrderKey] = input.HeaderOrder

	tlsReq.Header = headers

	return tlsReq, nil
}

func BuildResponse(sessionId string, withSession bool, resp *http.Response, cookies []*http.Cookie, isByteResponse bool) (Response, *TLSClientError) {
	defer resp.Body.Close()

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clientErr := NewTLSClientError(err)
		return Response{}, clientErr
	}

	finalResponse := string(respBodyBytes)

	if isByteResponse {
		mimeType := http.DetectContentType(respBodyBytes)
		base64Encoding := fmt.Sprintf("data:%s;base64,", mimeType)
		base64Encoding += base64.StdEncoding.EncodeToString(respBodyBytes)

		finalResponse = base64Encoding
	}

	response := Response{
		Status:  resp.StatusCode,
		Body:    finalResponse,
		Headers: resp.Header,
		Target:  "",
		Cookies: cookiesToMap(cookies),
	}

	if resp.Request != nil && resp.Request.URL != nil {
		response.Target = resp.Request.URL.String()
	}

	if withSession {
		response.SessionId = sessionId
	}

	return response, nil
}

func getTlsClient(requestInput RequestInput, sessionId string, withSession bool) (tls_client.HttpClient, error) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	tlsClientIdentifier := requestInput.TLSClientIdentifier
	proxyUrl := requestInput.ProxyUrl

	client, ok := clients[sessionId]

	if ok && withSession {
		modifiedClient, changed, err := handleModification(client, proxyUrl, requestInput.FollowRedirects)
		if err != nil {
			return nil, fmt.Errorf("failed to modify existing client: %w", err)
		}

		if changed {
			clients[sessionId] = modifiedClient
		}

		return modifiedClient, nil
	}

	var clientProfile tls_client.ClientProfile

	if requestInput.CustomTlsClient != nil {
		clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, err := getCustomTlsClientProfile(requestInput.CustomTlsClient)

		if err != nil {
			return nil, fmt.Errorf("can not build http client out of custom tls client information: %w", err)
		}

		clientProfile = tls_client.NewClientProfile(clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames)
	}

	if tlsClientIdentifier != "" {
		clientProfile = getTlsClientProfile(tlsClientIdentifier)
	}

	timeoutSeconds := tls_client.DefaultTimeoutSeconds

	if requestInput.TimeoutSeconds != 0 {
		timeoutSeconds = requestInput.TimeoutSeconds
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(timeoutSeconds),
		tls_client.WithClientProfile(clientProfile),
	}

	if !requestInput.WithoutCookieJar {
		jar, err := cookiejar.New(nil)

		if err != nil {
			return nil, fmt.Errorf("failed to build cookiejar")
		}

		options = append(options, tls_client.WithCookieJar(jar))
	}

	if !requestInput.FollowRedirects {
		options = append(options, tls_client.WithNotFollowRedirects())
	}

	if requestInput.InsecureSkipVerify {
		options = append(options, tls_client.WithInsecureSkipVerify())
	}

	proxy := proxyUrl

	if proxy != nil && *proxy != "" {
		options = append(options, tls_client.WithProxyUrl(*proxy))
	}

	tlsClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	if withSession {
		clients[sessionId] = tlsClient
	}

	return tlsClient, err
}

func getCustomTlsClientProfile(customClientDefinition *CustomTlsClient) (tls.ClientHelloID, map[http2.SettingID]uint32, []http2.SettingID, []string, uint32, []http2.Priority, error) {
	specFactory, err := tls_client.GetSpecFactoryFromJa3String(customClientDefinition.Ja3String, customClientDefinition.SupportedSignatureAlgorithms, customClientDefinition.SupportedVersions, customClientDefinition.KeyShareCurves, customClientDefinition.CertCompressionAlgo)

	if err != nil {
		return tls.ClientHelloID{}, nil, nil, nil, 0, nil, err
	}

	resolvedH2Settings := make(map[http2.SettingID]uint32)
	for key, value := range customClientDefinition.H2Settings {
		resolvedKey, ok := tls_client.H2SettingsMap[key]
		if !ok {
			continue
		}

		resolvedH2Settings[resolvedKey] = value
	}

	var resolvedH2SettingsOrder []http2.SettingID
	for _, order := range customClientDefinition.H2SettingsOrder {
		resolvedKey, ok := tls_client.H2SettingsMap[order]
		if !ok {
			continue
		}

		resolvedH2SettingsOrder = append(resolvedH2SettingsOrder, resolvedKey)
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

	return clientHelloId, resolvedH2Settings, resolvedH2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, nil
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

	if client == nil {
		return client, false, fmt.Errorf("no tls client for modification check")
	}

	if proxyUrl != nil {
		if client.GetProxy() != *proxyUrl {
			err := client.SetProxy(*proxyUrl)
			if err != nil {
				return nil, false, fmt.Errorf("failed to change proxy url of client: %w", err)
			}

			changed = true
		}
	}

	if client.GetFollowRedirect() != followRedirects {
		client.SetFollowRedirect(followRedirects)
		changed = true
	}

	return client, changed, nil
}

func cookiesToMap(cookies []*http.Cookie) map[string]string {
	ret := make(map[string]string, 0)

	for _, c := range cookies {
		ret[c.Name] = c.String()
	}

	return ret
}

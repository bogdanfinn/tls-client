// tls_client_cffi_src provides and manages a CFFI (C Foreign Function Interface) which allows code in other languages to interact with the module.
package tls_client_cffi_src

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/bogdanfinn/tls-client/profiles"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/tls-client"
	tls "github.com/bogdanfinn/utls"
	"github.com/google/uuid"
)

var clientsLock = sync.Mutex{}

// clients contains all registered clients, mapped by their individual IDs
var clients = make(map[string]tls_client.HttpClient)

// RemoveSession deletes the client with the given sessionId from the client session storage.
func RemoveSession(sessionId string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()
	client, ok := clients[sessionId]
	if !ok {
		return
	}
	client.CloseIdleConnections()

	delete(clients, sessionId)
}

// ClearSessionCache empties the client session storage.
func ClearSessionCache() {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// the remaining clients will be cleaned up by the garbage collection
	clients = make(map[string]tls_client.HttpClient)
}

// GetClient returns the client with the given sessionId from the client session storage.
// If there is no client with the given sessionId, it returns an error.
func GetClient(sessionId string) (tls_client.HttpClient, error) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	client, ok := clients[sessionId]

	if !ok {
		return nil, fmt.Errorf("no client found for sessionId: %s", sessionId)
	}

	return client, nil
}

// CreateClient creates a new client from a given RequestInput.
//
// The RequestInput should only contain a TLSClientIdentifier or a CustomTlsClient. If both are provided, an error will be returned.
func CreateClient(requestInput RequestInput) (client tls_client.HttpClient, sessionID string, withSession bool, clientErr *TLSClientError) {
	useSession := true
	sessionId := requestInput.SessionId

	newSessionId := uuid.New().String()
	if sessionId != nil && *sessionId != "" {
		newSessionId = *sessionId
	} else {
		useSession = false
	}

	if requestInput.TLSClientIdentifier != "" && requestInput.CustomTlsClient != nil {
		clientErr := NewTLSClientError(fmt.Errorf("cannot build client out of client identifier and custom tls client information. Please provide only one of them"))

		return nil, newSessionId, useSession, clientErr
	}

	if requestInput.TimeoutSeconds != 0 && requestInput.TimeoutMilliseconds != 0 {
		clientErr := NewTLSClientError(fmt.Errorf("cannot build client with both defined timeout in seconds and timeout in milliseconds. Please provide only one of them"))

		return nil, newSessionId, useSession, clientErr
	}

	tlsClient, err := getTlsClient(requestInput, newSessionId, useSession)
	if err != nil {
		clientErr := NewTLSClientError(fmt.Errorf("failed to build client out of request input: %w", err))

		return nil, newSessionId, useSession, clientErr
	}

	return tlsClient, newSessionId, useSession, nil
}

// BuildRequest constructs a HTTP request from a given RequestInput.
func BuildRequest(input RequestInput) (*http.Request, *TLSClientError) {
	var tlsReq *http.Request
	var err error

	if input.RequestMethod == "" || input.RequestUrl == "" {
		return nil, NewTLSClientError(fmt.Errorf("no request url or request method provided"))
	}

	if input.RequestBody != nil && *input.RequestBody != "" {
		requestBodyString := []byte(*input.RequestBody)
		if input.IsByteRequest {
			requestBodyString, err = base64.StdEncoding.DecodeString(*input.RequestBody)

			if err != nil {
				return nil, NewTLSClientError(fmt.Errorf("failed to base64 decode request body: %w", err))
			}
		}

		requestBody := bytes.NewBuffer(requestBodyString)
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

func readAllBodyWithStreamToFile(respBody io.ReadCloser, input RequestInput) ([]byte, error) {
	var respBodyBytes []byte
	var err error

	f, err := os.OpenFile(*input.StreamOutputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	blockSize := 1024 // 1 KB
	if input.StreamOutputBlockSize != nil {
		blockSize = *input.StreamOutputBlockSize
	}
	buf := make([]byte, blockSize)
	// Read the response body
	for {
		n, err := respBody.Read(buf)
		if err == io.EOF {
			if input.StreamOutputEOFSymbol != nil {
				f.Write([]byte(*input.StreamOutputEOFSymbol))
			}

			break
		}

		respBodyBytes = append(respBodyBytes, buf[:n]...)
		if _, err = f.Write(buf[:n]); err != nil {
			if input.WithDebug {
				fmt.Printf("Append stream output error: %+v\n", err)
			}

			return nil, err
		}

		if input.WithDebug {
			fmt.Printf("[stream decode result]==========\n%+v\n==========\n", string(buf[:n]))
		}
	}

	return respBodyBytes, nil
}

// BuildResponse constructs a client response from a given HTTP response. The client response can then be sent to the interface consumer.
func BuildResponse(sessionId string, withSession bool, resp *http.Response, cookies []*http.Cookie, input RequestInput) (Response, *TLSClientError) {
	defer resp.Body.Close()

	isByteResponse := input.IsByteResponse

	ce := resp.Header.Get("Content-Encoding")

	var respBodyBytes []byte
	var err error

	if !resp.Uncompressed {
		resp.Body = http.DecompressBodyByType(resp.Body, ce)
	}

	if input.StreamOutputPath != nil {
		respBodyBytes, err = readAllBodyWithStreamToFile(resp.Body, input)
	} else {
		respBodyBytes, err = io.ReadAll(resp.Body)
	}

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
		Id:           uuid.New().String(),
		Status:       resp.StatusCode,
		UsedProtocol: resp.Proto,
		Body:         finalResponse,
		Headers:      resp.Header,
		Target:       "",
		Cookies:      cookiesToMap(cookies),
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
		modifiedClient, changed, err := handleModification(client, proxyUrl, requestInput.FollowRedirects, requestInput.IsRotatingProxy)
		if err != nil {
			return nil, fmt.Errorf("failed to modify existing client: %w", err)
		}

		if changed {
			clients[sessionId] = modifiedClient
		}

		return modifiedClient, nil
	}

	clientProfile := profiles.DefaultClientProfile

	if requestInput.CustomTlsClient != nil {
		clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, headerPriority, err := getCustomTlsClientProfile(requestInput.CustomTlsClient)
		if err != nil {
			return nil, fmt.Errorf("can not build http client out of custom tls client information: %w", err)
		}

		clientProfile = profiles.NewClientProfile(clientHelloId, h2Settings, h2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, headerPriority)
	}

	if tlsClientIdentifier != "" {
		clientProfile = getTlsClientProfile(tlsClientIdentifier)
	}

	timeoutOption := tls_client.WithTimeoutSeconds(tls_client.DefaultTimeoutSeconds)

	if requestInput.TimeoutSeconds != 0 {
		timeoutOption = tls_client.WithTimeoutSeconds(requestInput.TimeoutSeconds)
	}

	if requestInput.TimeoutMilliseconds != 0 {
		timeoutOption = tls_client.WithTimeoutMilliseconds(requestInput.TimeoutMilliseconds)
	}

	options := []tls_client.HttpClientOption{
		timeoutOption,
		tls_client.WithClientProfile(clientProfile),
	}

	if requestInput.WithRandomTLSExtensionOrder {
		options = append(options, tls_client.WithRandomTLSExtensionOrder())
	}

	if requestInput.ForceHttp1 {
		options = append(options, tls_client.WithForceHttp1())
	}

	if requestInput.DisableIPV6 {
		options = append(options, tls_client.WithDisableIPV6())
	}

	if requestInput.TransportOptions != nil {
		transportOptions := &tls_client.TransportOptions{
			DisableKeepAlives:      requestInput.TransportOptions.DisableKeepAlives,
			DisableCompression:     requestInput.TransportOptions.DisableCompression,
			MaxIdleConns:           requestInput.TransportOptions.MaxIdleConns,
			MaxIdleConnsPerHost:    requestInput.TransportOptions.MaxIdleConnsPerHost,
			MaxConnsPerHost:        requestInput.TransportOptions.MaxConnsPerHost,
			MaxResponseHeaderBytes: requestInput.TransportOptions.MaxResponseHeaderBytes,
			WriteBufferSize:        requestInput.TransportOptions.WriteBufferSize,
			ReadBufferSize:         requestInput.TransportOptions.ReadBufferSize,
			IdleConnTimeout:        requestInput.TransportOptions.IdleConnTimeout,
			// RootCAs:                requestInput.TransportOptions.RootCAs,
		}

		options = append(options, tls_client.WithTransportOptions(transportOptions))
	}

	if requestInput.LocalAddress != nil {
		localAddr, err := net.ResolveTCPAddr("", *requestInput.LocalAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tcp address from local %s address: %w", *requestInput.LocalAddress, err)
		}

		options = append(options, tls_client.WithLocalAddr(*localAddr))
	}

	if requestInput.CatchPanics {
		options = append(options, tls_client.WithCatchPanics())
	}

	if len(requestInput.CertificatePinningHosts) > 0 {
		options = append(options, tls_client.WithCertificatePinning(requestInput.CertificatePinningHosts, nil))
	}

	if requestInput.WithDebug {
		options = append(options, tls_client.WithDebug())
	}

	if !requestInput.WithoutCookieJar {
		var jarOptions []tls_client.CookieJarOption
		if requestInput.WithDebug {
			jarOptions = append(jarOptions, tls_client.WithDebugLogger())
		}

		jar := tls_client.NewCookieJar(jarOptions...)

		if requestInput.WithDefaultCookieJar {
			jar, _ := cookiejar.New(nil)
			options = append(options, tls_client.WithCookieJar(jar))
		} else {
			options = append(options, tls_client.WithCookieJar(jar))
		}
	}

	if !requestInput.FollowRedirects {
		options = append(options, tls_client.WithNotFollowRedirects())
	}

	if requestInput.InsecureSkipVerify {
		options = append(options, tls_client.WithInsecureSkipVerify())
	}

	if requestInput.DefaultHeaders != nil && len(requestInput.DefaultHeaders) != 0 {
		options = append(options, tls_client.WithDefaultHeaders(requestInput.DefaultHeaders))
	}

	if requestInput.ServerNameOverwrite != nil && *requestInput.ServerNameOverwrite != "" {
		options = append(options, tls_client.WithServerNameOverwrite(*requestInput.ServerNameOverwrite))
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

func getCustomTlsClientProfile(customClientDefinition *CustomTlsClient) (tls.ClientHelloID, map[http2.SettingID]uint32, []http2.SettingID, []string, uint32, []http2.Priority, *http2.PriorityParam, error) {
	specFactory, err := tls_client.GetSpecFactoryFromJa3String(customClientDefinition.Ja3String, customClientDefinition.SupportedSignatureAlgorithms, customClientDefinition.SupportedDelegatedCredentialsAlgorithms, customClientDefinition.SupportedVersions, customClientDefinition.KeyShareCurves, customClientDefinition.ALPNProtocols, customClientDefinition.ALPSProtocols, customClientDefinition.ECHCandidateCipherSuites.Translate(), customClientDefinition.ECHCandidatePayloads, customClientDefinition.CertCompressionAlgo)
	if err != nil {
		return tls.ClientHelloID{}, nil, nil, nil, 0, nil, nil, err
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

	var headerPriority *http2.PriorityParam

	if customClientDefinition.HeaderPriority != nil {
		headerPriority = &http2.PriorityParam{
			StreamDep: customClientDefinition.HeaderPriority.StreamDep,
			Exclusive: customClientDefinition.HeaderPriority.Exclusive,
			Weight:    customClientDefinition.HeaderPriority.Weight,
		}
	}

	clientHelloId := tls.ClientHelloID{
		Client:      "Custom",
		Version:     "1",
		Seed:        nil,
		SpecFactory: specFactory,
	}

	return clientHelloId, resolvedH2Settings, resolvedH2SettingsOrder, pseudoHeaderOrder, connectionFlow, priorityFrames, headerPriority, nil
}

func getTlsClientProfile(tlsClientIdentifier string) profiles.ClientProfile {
	tlsClientProfile, ok := profiles.MappedTLSClients[tlsClientIdentifier]

	if !ok {
		return profiles.DefaultClientProfile
	}

	return tlsClientProfile
}

func handleModification(client tls_client.HttpClient, proxyUrl *string, followRedirects bool, isRotatingProxy bool) (tls_client.HttpClient, bool, error) {
	changed := false

	if client == nil {
		return client, false, fmt.Errorf("no tls client for modification check")
	}

	if proxyUrl != nil {
		if client.GetProxy() != *proxyUrl || isRotatingProxy {
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
		ret[c.Name] = c.Value
	}

	return ret
}

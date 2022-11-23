package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"net/url"
	"unsafe"

	http "github.com/bogdanfinn/fhttp"
	tls_client_cffi_src "github.com/bogdanfinn/tls-client/cffi_src"
)

//export freeAll
func freeAll() *C.char {
	err := tls_client_cffi_src.DestroyTlsClientSessions()

	if err != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(err)
		return handleErrorResponse("", false, clientErr)
	}

	out := tls_client_cffi_src.FreeOutput{
		Success: true,
	}

	jsonResponse, marshallError := json.Marshal(out)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse("", false, clientErr)
	}

	responseString := C.CString(string(jsonResponse))

	defer C.free(unsafe.Pointer(responseString))
	return responseString
}

//export freeSession
func freeSession(freeSessionParams *C.char) *C.char {
	freeSessionParamsJson := C.GoString(freeSessionParams)

	freeSessionInput := tls_client_cffi_src.FreeSessionInput{}
	marshallError := json.Unmarshal([]byte(freeSessionParamsJson), &freeSessionInput)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse("", false, clientErr)
	}

	err := tls_client_cffi_src.DestroyTlsClientSession(freeSessionInput.SessionId)

	if err != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(err)
		return handleErrorResponse(freeSessionInput.SessionId, true, clientErr)
	}

	out := tls_client_cffi_src.FreeOutput{
		Success: true,
	}

	jsonResponse, marshallError := json.Marshal(out)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse(freeSessionInput.SessionId, true, clientErr)
	}

	responseString := C.CString(string(jsonResponse))

	defer C.free(unsafe.Pointer(responseString))
	return responseString
}

//export getCookiesFromSession
func getCookiesFromSession(getCookiesParams *C.char) *C.char {
	getCookiesParamsJson := C.GoString(getCookiesParams)

	cookiesInput := tls_client_cffi_src.GetCookiesFromSessionInput{}
	marshallError := json.Unmarshal([]byte(getCookiesParamsJson), &cookiesInput)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse("", false, clientErr)
	}

	tlsClient, err := tls_client_cffi_src.GetTlsClientFromSession(cookiesInput.SessionId)

	if err != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(err)
		return handleErrorResponse(cookiesInput.SessionId, true, clientErr)
	}

	u, parsErr := url.Parse(cookiesInput.Url)
	if parsErr != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(parsErr)
		return handleErrorResponse(cookiesInput.SessionId, true, clientErr)
	}

	cookies := tlsClient.GetCookies(u)

	jsonResponse, marshallError := json.Marshal(cookies)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse(cookiesInput.SessionId, true, clientErr)
	}

	responseString := C.CString(string(jsonResponse))

	defer C.free(unsafe.Pointer(responseString))
	return responseString
}

//export request
func request(requestParams *C.char) *C.char {
	requestParamsJson := C.GoString(requestParams)

	requestInput := tls_client_cffi_src.RequestInput{}
	marshallError := json.Unmarshal([]byte(requestParamsJson), &requestInput)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse("", false, clientErr)
	}

	tlsClient, sessionId, withSession, err := tls_client_cffi_src.GetTlsClientFromInput(requestInput)

	if err != nil {
		return handleErrorResponse(sessionId, withSession, err)
	}

	req, err := tls_client_cffi_src.BuildRequest(requestInput)

	if err != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(err)
		return handleErrorResponse(sessionId, withSession, clientErr)
	}

	cookies := buildCookies(requestInput.RequestCookies)

	if len(cookies) > 0 {
		tlsClient.SetCookies(req.URL, cookies)
	}

	resp, reqErr := tlsClient.Do(req)

	if reqErr != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(fmt.Errorf("failed to do request: %w", reqErr))
		return handleErrorResponse(sessionId, withSession, clientErr)
	}

	sessionCookies := tlsClient.GetCookies(req.URL)

	response, err := tls_client_cffi_src.BuildResponse(sessionId, withSession, resp, sessionCookies, requestInput.IsByteResponse)
	if err != nil {
		return handleErrorResponse(sessionId, withSession, err)
	}

	jsonResponse, marshallError := json.Marshal(response)

	if marshallError != nil {
		clientErr := tls_client_cffi_src.NewTLSClientError(marshallError)
		return handleErrorResponse(sessionId, withSession, clientErr)
	}

	responseString := C.CString(string(jsonResponse))

	defer C.free(unsafe.Pointer(responseString))
	return responseString
}

func handleErrorResponse(sessionId string, withSession bool, err *tls_client_cffi_src.TLSClientError) *C.char {
	response := tls_client_cffi_src.Response{
		Status:  0,
		Body:    err.Error(),
		Headers: nil,
		Cookies: nil,
	}

	if withSession {
		response.SessionId = sessionId
	}

	jsonResponse, marshallError := json.Marshal(response)

	if marshallError != nil {
		errStr := C.CString(marshallError.Error())
		defer C.free(unsafe.Pointer(errStr))

		return errStr
	}

	responseString := C.CString(string(jsonResponse))

	defer C.free(unsafe.Pointer(responseString))
	return responseString
}

func buildCookies(cookies []tls_client_cffi_src.CookieInput) []*http.Cookie {
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

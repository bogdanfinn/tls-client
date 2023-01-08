package hawk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	lz "github.com/Lazarus/lz-string-go"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

func solveFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// get initial response body
		if len(s.initialResponseBody) != 0 && len(s.urlPart) != 0 && len(s.keyStrUriSafe) != 0 {
			return s, nil
		}

		extractedScript, err := extractChallengeScript(string(s.originalResponseBody))
		if err != nil {
			return s, err
		}

		// Fetching CF script
		script := fmt.Sprintf("https://%s%s", s.domain, *extractedScript)

		req, err := http.NewRequest(http.MethodGet, script, nil)
		if err != nil {
			return s, fmt.Errorf("could no create request for challenging solving: %w", err)
		}

		req.Header = initHeaders
		req.Header["user-agent"] = s.originalResponse.Request.Header["user-agent"]
		resp, err := client.Do(req)
		if err != nil {
			return s, fmt.Errorf("could not request: %w", err)
		}

		defer resp.Body.Close()

		initialResponseBody, err := readAndCloseBody(resp.Body)
		if err != nil {
			return s, fmt.Errorf("could not request: %w", err)
		}

		urlPartP, err := extractUrlPartRegex(string(s.initialResponseBody))
		if err != nil {
			return s, err
		}
		urlPart := *urlPartP

		keyStringUri, err := extractKeyStringUri(string(s.initialResponseBody))
		if err != nil {
			return s, err
		}

		s.urlPart = urlPart
		s.keyStrUriSafe = *keyStringUri

		s.initialResponseBody = initialResponseBody

		logger.Debug("Loaded init script.")

		return s, nil
	}
}

// initUrl
// initURL
// requestURL
// result
// name
// baseObj
// requestPass
// requestR
// ts
// md
func challengeInitiationPayloadFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		if len(s.initURL) != 0 && len(s.requestURL) != 0 && len(s.result) != 0 && len(s.name) != 0 && len(s.baseObj) != 0 && len(s.requestPass) != 0 && len(s.requestR) != 0 && s.ts != 0 && len(s.md) != 0 {
			return s, nil
		}

		payload, err := json.Marshal(map[string]interface{}{
			"body":    base64.StdEncoding.EncodeToString(s.originalResponseBody),
			"url":     s.urlPart,
			"domain":  s.domain,
			"captcha": config.Captcha,
			"key":     s.keyStrUriSafe,
		})

		if err != nil {
			return s, fmt.Errorf("could not marshal body: %w", err)
		}

		requestUrl, err := url.Parse(fmt.Sprintf("https://%s/cf-a/ov1/p1", config.ApiDomain))
		if err != nil {
			return s, fmt.Errorf("could not create requestUrl: %w", err)
		}

		for key, value := range config.AuthParams {
			requestUrl.Query().Set(key, value)
		}

		challengePayload, err := client.Post(requestUrl.String(), "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return s, err
		}
		defer challengePayload.Body.Close()

		var challengePayloadResponse apiResponse
		err = readAndUnmarshalBody(challengePayload.Body, &challengePayloadResponse)
		if err != nil {
			return s, err
		}

		s.initURL = challengePayloadResponse.URL
		s.requestURL = challengePayloadResponse.ResultURL
		s.result = challengePayloadResponse.Result
		s.name = challengePayloadResponse.Name
		s.baseObj = challengePayloadResponse.BaseObj
		s.requestPass = challengePayloadResponse.Pass
		s.requestR = challengePayloadResponse.R
		s.ts = challengePayloadResponse.TS
		s.md = challengePayloadResponse.Md

		logger.Debug("Submitted init payload to the api.")

		return s, nil
	}
}

func initiateCloudflareFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// challengePayloadBody
		if len(s.challengePayloadBody) != 0 {
			return s, nil
		}

		resultDecoded, err := base64.StdEncoding.DecodeString(s.result)
		if err != nil {
			return s, err
		}

		payload := createParams(map[string]string{
			s.name: lz.Compress(string(resultDecoded), s.keyStrUriSafe),
		})

		if err != nil {
			return s, err
		}

		req, err := http.NewRequest(http.MethodPost, s.initURL, bytes.NewBufferString(payload))
		if err != nil {
			return s, err
		}

		req.Header = challengeHeaders
		initURLSplit := strings.Split(s.initURL, "/")
		req.Header["cf-challenge"] = []string{initURLSplit[len(initURLSplit)-1]}
		req.Header["referer"] = []string{strings.Split(s.originalResponse.Request.URL.String(), "?")[0]}
		req.Header["origin"] = []string{"https://" + s.domain}
		req.Header["user-agent"] = s.originalResponse.Request.Header["user-agent"]
		resp, err := client.Do(req)
		if err != nil {
			return s, err
		}

		challengePayloadBody, err := readAndCopyBody(resp)
		if err != nil {
			return s, err
		}
		defer resp.Body.Close()

		s.challengePayloadBody = string(challengePayloadBody)
		s.challengePayloadBody = string(challengePayloadBody)

		logger.Debug("Initiated challenge.")

		return s, nil
	}
}

func solvePayloadFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// result
		// Fetches main challenge payload from hawk api

		body := map[string]interface{}{
			"body_home":   base64.RawURLEncoding.EncodeToString(s.originalResponseBody),
			"body_sensor": base64.RawURLEncoding.EncodeToString([]byte(s.challengePayloadBody)),
			"result":      s.baseObj,
			"ts":          s.ts,
			"url":         s.initURL,
		}

		if len(s.result) != 0 {
			body["rerun_base"] = s.result
			body["rerun"] = true
		} else {
			body["ua"] = s.originalResponse.Request.Header["user-agent"][0]
		}

		payload, err := json.Marshal(map[string]interface{}{
			"body_home":   base64.RawURLEncoding.EncodeToString(s.originalResponseBody),
			"body_sensor": base64.RawURLEncoding.EncodeToString([]byte(s.challengePayloadBody)),
			"result":      s.baseObj,
			"ts":          s.ts,
			"url":         s.initURL,
			"ua":          s.originalResponse.Request.Header["user-agent"][0],
		})
		if err != nil {
			return s, err
		}

		requestUrl, err := url.Parse(fmt.Sprintf("https://%s/cf-a/ov1/p2", config.ApiDomain))
		if err != nil {
			return s, err
		}

		for key, value := range config.AuthParams {
			requestUrl.Query().Set(key, value)
		}

		cc, err := client.Post(requestUrl.String(), "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return s, err
		}
		defer cc.Body.Close()

		var solvePayloadResponse apiResponse
		if err := readAndUnmarshalBody(cc.Body, &solvePayloadResponse); err != nil {
			return s, err
		}
		s.result = solvePayloadResponse.Result
		s.result = solvePayloadResponse.Result

		logger.Debug("Fetched challenge payload.")

		return s, nil
	}
}

func sendMainPayloadFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// mainPayloadResponseBody
		// Sends the main payload to cf
		if len(s.mainPayloadResponseBody) != 0 {
			return s, nil
		}

		resultDecoded, err := base64.StdEncoding.DecodeString(s.result)
		if err != nil {
			return s, err
		}
		payload := createParams(map[string]string{
			s.name: lz.Compress(string(resultDecoded), s.keyStrUriSafe),
		})

		if err != nil {
			return s, err
		}

		req, err := http.NewRequest(http.MethodPost, s.initURL, bytes.NewBufferString(payload))
		if err != nil {
			return s, err
		}

		req.Header = challengeHeaders

		initURLSplit := strings.Split(s.initURL, "/")
		req.Header["cf-challenge"] = []string{initURLSplit[len(initURLSplit)-1]}
		req.Header["referer"] = []string{strings.Split(s.originalResponse.Request.URL.String(), "?")[0]}
		req.Header["origin"] = []string{"https://" + s.domain}
		req.Header["user-agent"] = s.originalResponse.Request.Header["user-agent"]

		mainPayloadResponse, err := client.Do(req)
		if err != nil {
			return s, err
		}

		body, err := readAndCopyBody(mainPayloadResponse)
		s.mainPayloadResponseBody = string(body)

		defer mainPayloadResponse.Body.Close()

		logger.Debug("Submitted challenge.")

		return s, nil
	}
}

func getChallengeResultFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// get finalApi
		if s.finalApi != nil {
			return s, nil
		}

		payload, err := json.Marshal(map[string]interface{}{
			"body_sensor": base64.StdEncoding.EncodeToString([]byte(s.mainPayloadResponseBody)),
			"result":      s.baseObj,
		})
		if err != nil {
			return s, err
		}

		requestUrl, err := url.Parse(fmt.Sprintf("https://%s/cf-a/ov1/p3", config.ApiDomain))
		if err != nil {
			return s, err
		}

		for key, value := range config.AuthParams {
			requestUrl.Query().Set(key, value)
		}

		ee, err := client.Post(requestUrl.String(), "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return s, err
		}
		defer ee.Body.Close()

		response := &apiResponse{}
		if err = readAndUnmarshalBody(ee.Body, response); err != nil {
			return s, err
		}

		s.finalApi = response

		logger.Debug("Fetched challenge response.")

		return s, nil
	}
}

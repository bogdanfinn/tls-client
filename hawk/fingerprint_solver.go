package hawk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

// get initialResponseBody
func initiateScriptFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		if len(s.initialResponseBody) != 0 {
			return s, nil
		}

		urlPath := strings.Split(strings.Split(string(s.originalResponseBody), `<script src="`)[3], `"`)[0]
		s.initURL = fmt.Sprintf("https://%s%s", s.domain, urlPath)

		req, err := http.NewRequest(http.MethodGet, s.initURL, nil)
		if err != nil {
			return s, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return s, err
		}
		defer resp.Body.Close()

		initialResponseBody, err := readAndCloseBody(resp.Body)
		if err != nil {
			return s, err
		}

		s.initialResponseBody = initialResponseBody

		return s, nil
	}
}

// get p1 result/url
func getPayloadFromAPIFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		/*
				  Recieve the needed fingerprint data from hawk api
			        :return:
		*/

		if len(s.p1Result) != 0 && len(s.p1TargetURL) != 0 {
			return s, nil
		}

		payload, err := json.Marshal(map[string]string{
			"body": base64.StdEncoding.EncodeToString(s.initialResponseBody),
			"url":  s.initURL,
		})
		if err != nil {
			return s, err
		}

		requestUrl, err := url.Parse(fmt.Sprintf("https://%s/cf-a/fp/p1", config.ApiDomain))
		if err != nil {
			return s, err
		}

		for key, value := range config.AuthParams {
			requestUrl.Query().Set(key, value)
		}

		resp, err := client.Post(requestUrl.String(), "application/json", bytes.NewReader(payload))
		if err != nil {
			return s, err
		}
		defer resp.Body.Close()

		var p1Response apiResponse
		if err := readAndUnmarshalBody(resp.Body, &p1Response); err != nil {
			return s, err
		}
		s.p1Result = p1Response.Result
		s.p1TargetURL = p1Response.URL

		return s, nil
	}
}

func submitFingerprintChallengeFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		/*
		   Submit the fingerprint data to cloudflare
		          :return:
		*/

		logger.Debug("Submitting fingerprint")
		result, err := client.Post(s.p1TargetURL, "", bytes.NewBufferString(s.result))
		if err != nil {
			return s, err
		}
		defer result.Body.Close()

		if result.StatusCode == http.StatusTooManyRequests {
			return s, fmt.Errorf("FP DATA declined")
		}
		if result.StatusCode == http.StatusNotFound {
			return s, errors.New("Fp ep changed")
		}

		return s, nil
	}
}

// get final response
func getPageFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		/*
				 Perform the original request
			        :return:
		*/
		if s.finalResponse != nil {
			return s, nil
		}

		logger.Debug("Fetching original request target")

		url := s.originalResponse.Request.URL.String()

		if strings.Contains(s.originalResponse.Request.URL.String(), "?") {
			url = strings.Split(s.originalResponse.Request.URL.String(), "?")[0]
		}

		finalResp, err := client.Get(url)
		if err != nil {
			return s, err
		}
		s.finalResponse = finalResp

		return s, nil
	}
}

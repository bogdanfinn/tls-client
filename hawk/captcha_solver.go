package hawk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	lz "github.com/Lazarus/lz-string-go"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

// In Progress
func handleFirstCaptchaFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		/* Handling captcha
		   Note that this function is designed to work with cloudscraper,
		   if you are building your own flow you will need to rework this part a bit.
		*/

		var token string
		var err error
		if s.finalApi.Click {
			token = "click"
		} else {
			logger.Debug("Captcha needed, requesting token.")

			token, err = config.CaptchaFunc(s.originalResponse.Request.URL.String(), s.finalApi.SiteKey)
			if err != nil {
				return s, err
			}
		}

		payload, err := json.Marshal(map[string]interface{}{
			"result":             s.result,
			"token":              token,
			"h-captcha-response": token,
			"data":               s.finalApi.Result,
		})
		if err != nil {
			return s, err
		}

		requestUrl, err := url.Parse(fmt.Sprintf("https://%v/cf-a/ov1/cap1", config.ApiDomain))
		if err != nil {
			return s, err
		}

		for key, value := range config.AuthParams {
			requestUrl.Query().Set(key, value)
		}

		ff, err := client.Post(requestUrl.String(), "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return s, err
		}
		defer ff.Body.Close()

		var handleCaptchaResponse apiResponse
		if err := readAndUnmarshalBody(ff.Body, &handleCaptchaResponse); err != nil {
			return s, err
		}
		s.firstCaptchaResult = handleCaptchaResponse

		return s, nil
	}

}

func submitCaptchaFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// Submits the challenge + captcha and trys to access target url
		if !s.captchaResponseAPI.Valid {
			return s, nil
		}

		payloadMap := map[string]string{
			"r":               s.requestR,
			"cf_captcha_kind": "h",
			"vc":              s.requestPass,
			"captcha_vc":      s.captchaResponseAPI.JschlVc,
			"captcha_answer":  s.captchaResponseAPI.JschlAnswer,
			"cf_ch_verify":    "plat",
		}

		if s.captchaResponseAPI.CfChCpReturn != "" {
			payloadMap["cf_ch_cp_return"] = s.captchaResponseAPI.CfChCpReturn
		}

		if s.md != "" {
			payloadMap["md"] = s.md
		}

		// "captchka" Spelling mistake?
		payloadMap["h-captcha-response"] = "captchka"

		payload := createParams(payloadMap)

		req, err := http.NewRequest(http.MethodPost, s.requestURL, bytes.NewBufferString(payload))
		if err != nil {
			return s, err
		}

		req.Header = submitHeaders
		req.Header["referer"] = []string{s.originalResponse.Request.URL.String()}
		req.Header["origin"] = []string{"https://" + s.domain}
		req.Header["user-agent"] = s.originalResponse.Request.Header["user-agent"]

		if (time.Now().Unix() - s.startTime.Unix()) < 5 {
			// Waiting X amount of sec for CF delay
			logger.Debug("sleeping %d sec for cf delay", 5-(time.Now().Unix()-s.startTime.Unix()))

			time.Sleep(time.Duration(5-(time.Now().Unix()-s.startTime.Unix())) * time.Second)
		}

		final, err := client.Do(req)
		if err != nil {
			return s, err
		}
		defer final.Body.Close()

		s.finalResponse = final

		logger.Debug("submitted captcha challange")

		return s, nil
	}
}

func handleSecondCaptchaFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		/* Handling captcha
		   Note that this function is designed to work with cloudscraper,
		   if you are building your own flow you will need to rework this part a bit.
		*/
		resultDecoded, err := base64.StdEncoding.DecodeString(s.firstCaptchaResult.Result)
		if err != nil {
			return s, fmt.Errorf("posting to cloudflare challenge endpoint error: %w", err)
		}

		payload := []byte(createParams(map[string]string{
			s.name: lz.Compress(string(resultDecoded), s.keyStrUriSafe),
		}))

		req, err := http.NewRequest(http.MethodPost, s.initURL, bytes.NewBuffer(payload))
		if err != nil {
			return s, err
		}

		req.Header = challengeHeaders
		initURLSplit := strings.Split(s.initURL, "/")
		req.Header["cf-challenge"] = []string{initURLSplit[len(initURLSplit)-1]}
		req.Header["referer"] = []string{strings.Split(s.originalResponse.Request.URL.String(), "?")[0]}
		req.Header["origin"] = []string{"https://" + s.domain}
		req.Header["user-agent"] = s.originalResponse.Request.Header["user-agent"]

		gg, err := client.Do(req)
		if err != nil {
			return s, err
		}
		defer gg.Body.Close()

		body, err := readAndCloseBody(gg.Body)
		if err != nil {
			return s, err
		}

		payload, err = json.Marshal(map[string]interface{}{
			"body_sensor": base64.StdEncoding.EncodeToString(body),
			"result":      s.baseObj,
		})
		if err != nil {
			return s, err
		}

		requestUrl, err := url.Parse(fmt.Sprintf("https://%v/cf-a/ov1/cap2", config.ApiDomain))
		if err != nil {
			return s, err
		}

		for key, value := range config.AuthParams {
			requestUrl.Query().Set(key, value)
		}

		hh, err := client.Post(requestUrl.String(), "application/json", bytes.NewBuffer(payload))
		if err != nil {
			return s, err
		}
		defer hh.Body.Close()

		handleCaptchaResponse := apiResponse{}
		err = readAndUnmarshalBody(hh.Body, &handleCaptchaResponse)
		if err != nil {
			return s, err
		}
		s.captchaResponseAPI = handleCaptchaResponse

		return s, fmt.Errorf("captcha was not accepted - most likly wrong token")
	}
}

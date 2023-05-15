package hawk

import (
	"fmt"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

type hawkCF struct {
	apiKey string
	config cfConfig
}

func NewHawkCF(apiKey string) (tls_client.CFSolvingHandler, error) {
	/*jar := tls_client.NewCookieJar()

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(60),
		tls_client.WithClientProfile(tls_client.CloudflareCustom),
		tls_client.WithCookieJar(jar),
	}

	httpClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	if err != nil {
		return nil, err
	}*/

	cfConfig := cfConfig{
		ApiDomain: "cf-v2.hwkapi.com",
		AuthParams: map[string]string{
			"auth": apiKey,
		},
		ErrorDelay: 30,
		MaxRetries: 5,
	}

	return &hawkCF{
		apiKey: apiKey,
		config: cfConfig,
	}, nil
}

func (h *hawkCF) Solve(logger tls_client.Logger, client tls_client.HttpClient, response *http.Response) (*http.Response, error) {
	originalResponseBody, err := readAndCopyBody(response)
	if err != nil {
		return response, err
	}

	challengeState := cfChallengeState{
		originalResponseBody: originalResponseBody,
		originalResponse:     response,
		startTime:            time.Now(),
		domain:               response.Request.URL.Host,
	}

	if isNewIUAMChallenge(response) {
		resp, err := h.run(logger, client, challengeState)
		if err != nil {
			return response, err
		}

		return resp, nil
	}

	if isNewCaptchaChallenge(response) {
		h.config.Captcha = true

		resp, err := h.run(logger, client, challengeState)
		if err != nil {
			return response, err
		}

		return resp, nil
	}

	if isFingerprintChallenge(response) {
		h.config.FingerPrint = true

		resp, err := h.fingerprint(logger, client, challengeState)
		if err != nil {
			return response, err
		}

		return resp, nil
	}

	return response, nil
}

func (h *hawkCF) run(logger tls_client.Logger, client tls_client.HttpClient, challengeState cfChallengeState) (*http.Response, error) {
	for retries := 0; challengeState.finalResponse == nil || h.config.MaxRetries >= retries; retries++ {
		if h.config.Captcha && h.config.CaptchaFunc == nil {
			return nil, fmt.Errorf("captcha is present with nil CaptchaFunction")
		}

		var err error
		for _, f := range getChallengeSolver(logger, client, h.config) {
			challengeState, err = executeWithRetries(f, h.config, challengeState)
			if err != nil {
				return nil, err
			}
		}

		if challengeState.finalApi.Status == "rerun" {
			continue
		}

		if !h.config.Captcha && challengeState.finalApi.Captcha {
			return nil, fmt.Errorf("cf returned captcha and captcha handling is disabled")
		}

		if challengeState.finalApi.Captcha {
			for _, f := range getCaptchaSolver(logger, client, h.config) {
				challengeState, err = executeWithRetries(f, h.config, challengeState)
				if err != nil {
					return nil, err
				}
			}

			continue
		}

		submitChallenge := submitChallengeFactory(logger, client, h.config)

		challengeState, err = executeWithRetries(submitChallenge, h.config, challengeState)
		if err != nil {
			return nil, err
		}
	}

	return challengeState.finalResponse, nil
}

func (h *hawkCF) fingerprint(logger tls_client.Logger, client tls_client.HttpClient, challengeState cfChallengeState) (*http.Response, error) {
	var err error

	for _, f := range getFingerprintSolver(logger, client, h.config) {
		challengeState, err = executeWithRetries(f, h.config, challengeState)
		if err != nil {
			return nil, err
		}
	}

	return challengeState.finalResponse, nil
}

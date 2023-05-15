package hawk

import (
	"bytes"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

type solverFactory func(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver
type solver func(challengeState cfChallengeState) (cfChallengeState, error)

var captchaSolver = []solverFactory{
	handleFirstCaptchaFactory,
	handleSecondCaptchaFactory,
	submitCaptchaFactory,
}

func getCaptchaSolver(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) []solver {
	var solver []solver

	for _, solverFactory := range captchaSolver {
		solver = append(solver, solverFactory(logger, client, config))
	}

	return solver
}

var fingerPrintSolver = []solverFactory{
	initiateScriptFactory,
	getPayloadFromAPIFactory,
	submitFingerprintChallengeFactory,
	getPageFactory,
}

func getFingerprintSolver(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) []solver {
	var solver []solver

	for _, solverFactory := range fingerPrintSolver {
		solver = append(solver, solverFactory(logger, client, config))
	}

	return solver
}

var challengeSolver = []solverFactory{
	solveFactory,
	challengeInitiationPayloadFactory,
	initiateCloudflareFactory,
	solvePayloadFactory,
	sendMainPayloadFactory,
	getChallengeResultFactory,
}

func getChallengeSolver(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) []solver {
	var solver []solver

	for _, solverFactory := range challengeSolver {
		solver = append(solver, solverFactory(logger, client, config))
	}

	return solver
}

func submitChallengeFactory(logger tls_client.Logger, client tls_client.HttpClient, config cfConfig) solver {
	return func(s cfChallengeState) (cfChallengeState, error) {
		// Submits the challenge and trys to access target url

		payloadMap := map[string]string{
			"r":            s.requestR,
			"jschl_vc":     s.finalApi.JschlVc,
			"pass":         s.requestPass,
			"jschl_answer": s.finalApi.JschlAnswer,
			"cf_ch_verify": "plat",
		}

		// cf added a new flow where they present a 503 followed up by a 403 captcha
		if s.finalApi.CfChCpReturn != "" {
			payloadMap["cf_ch_cp_return"] = s.finalApi.CfChCpReturn
		}

		if s.md != "" {
			payloadMap["md"] = s.md
		}
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
			logger.Debug("sleeping %v sec for cf delay", 5-(time.Now().Unix()-s.startTime.Unix()))

			time.Sleep(time.Duration(5-(time.Now().Unix()-s.startTime.Unix())) * time.Second)
		}
		final, err := client.Do(req)
		if err != nil {
			return s, err

		}
		defer final.Body.Close()

		logger.Debug("Submitted final challange.")

		if final.StatusCode != http.StatusForbidden {
			s.finalResponse = final

			return s, err
		}

		body, err := readAndCopyBody(final)
		if err != nil {
			return s, err
		}

		if !checkForCaptcha(string(body)) {
			s.finalResponse = final

			return s, nil
		}

		// as this was a 403 post we need to get again dont ask why just do it
		req, err = http.NewRequest(http.MethodGet, s.originalResponse.Request.URL.String(), nil)
		if err != nil {
			return s, err
		}

		req.Header = s.originalResponse.Request.Header
		weirdGetReq, err := client.Do(req)
		if err != nil {
			return s, err
		}
		defer weirdGetReq.Body.Close()

		config.Captcha = true
		s.originalResponse = weirdGetReq
		originalBodyResponse, err := readAndCloseBody(weirdGetReq.Body)
		if err != nil {
			return s, err
		}
		// we have to start again
		s.originalResponseBody = originalBodyResponse
		s.domain = s.originalResponse.Request.URL.Host

		return s, nil
	}
}

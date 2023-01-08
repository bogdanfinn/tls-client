package hawk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	http "github.com/bogdanfinn/fhttp"
)

var scriptReg = regexp.MustCompile("['|\"](/cdn-cgi/.+?)['|\"]")
var urlPartRegex = regexp.MustCompile(`0\.[^\[\]\{\}\(\) \'\|\/)]{40,}`)
var keyStringUriSaveRegex = regexp.MustCompile(`[\W]?([A-Za-z0-9+\-$]{65})[\W]`)

const challengePlatform = "challenge-platform"

func readAndCopyBody(r interface{}) ([]byte, error) {
	var bodyReadCloser io.ReadCloser
	switch r.(type) {
	case *http.Response:
		bodyReadCloser = r.(*http.Response).Body
	}

	defer bodyReadCloser.Close()

	var body []byte
	var err error
	var b bytes.Buffer
	t := io.TeeReader(bodyReadCloser, &b)
	body, err = io.ReadAll(t)
	if err != nil {
		return body, err
	}

	newReadCloser := io.NopCloser(bytes.NewBuffer(body))
	switch r.(type) {
	case *http.Response:
		r.(*http.Response).Body = newReadCloser
	}

	return body, err
}

func readAndUnmarshalBody(respBody io.ReadCloser, x interface{}) error {
	body, err := io.ReadAll(respBody)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &x)

	return err
}

func readAndCloseBody(respBody io.ReadCloser) ([]byte, error) {
	defer respBody.Close()

	return io.ReadAll(respBody)
}

func createParams(paramsLong map[string]string) string {
	params := url.Values{}
	for key, value := range paramsLong {
		params.Add(key, value)
	}

	return params.Encode()
}

func checkForCaptcha(body string) bool {
	doc := soup.HTMLParse(body)

	element := doc.Find("input", "name", "cf_captcha_kind")
	if element.Error != nil {
		return false
	}
	if val, ok := element.Attrs()["value"]; ok && val == "h" {
		return true
	}

	return false
}

func extractKeyStringUri(body string) (*string, error) {
	matches := keyStringUriSaveRegex.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		matchWithoutComma := strings.ReplaceAll(m[1], ",", "")
		if strings.Contains(matchWithoutComma, "+") && strings.Contains(matchWithoutComma, "-") && strings.Contains(matchWithoutComma, "$") && len(matchWithoutComma) != 0 {
			return &matchWithoutComma, nil
		}
	}

	return nil, fmt.Errorf("could not extract the key string uri from the repsonse body")

}

func extractUrlPartRegex(body string) (*string, error) {
	matches := urlPartRegex.FindAllStringSubmatch(body, 1)
	if len(matches) == 0 || len(matches[0]) == 0 {
		return nil, fmt.Errorf("could not extract the url part from the response body")
	}

	return &matches[0][1], nil
}

func executeWithRetries(f solver, config cfConfig, challengeState cfChallengeState) (cfChallengeState, error) {
	var err error

	for i := 0; i < config.MaxRetries; i++ {
		challengeState, err = f(challengeState)
		if err == nil {
			return challengeState, nil
		}
		if isNotRetryableError(err) {
			return challengeState, err
		}

		time.Sleep(config.ErrorDelay)
	}

	return challengeState, err
}

func extractChallengeScript(body string) (*string, error) {
	matches := scriptReg.FindAllStringSubmatch(body, -1)
	if matches == nil {
		return nil, fmt.Errorf("could not find a match for the script regex")
	}

	for _, m := range matches {
		if strings.Contains(m[1], challengePlatform) {
			match := m[1]

			return &match, nil
		}
	}

	return nil, fmt.Errorf("could not find a match for the script regex with 'challenge-platform' inside")

}

func isNewIUAMChallenge(response *http.Response) bool {
	body, err := readAndCopyBody(response)
	if err != nil {
		return false
	}
	firstReg, err := regexp.MatchString(`cpo.src\s*=\s*[",']/cdn-cgi/challenge-platform/?\w?/?\w?/orchestrate/jsch/v1`, string(body))
	if err != nil {
		return false
	}
	secondReg, err := regexp.MatchString(`window._cf_chl_opt`, string(body))
	if err != nil {
		return false
	}

	return strings.Contains(response.Header.Get("Server"), "cloudflare") &&
		(response.StatusCode == 429 || response.StatusCode == 403 || response.StatusCode == 503) &&
		firstReg && secondReg

}

func isFingerprintChallenge(response *http.Response) bool {
	if response.StatusCode == 429 {
		body, err := readAndCopyBody(response)
		if err != nil {
			return false
		}
		if strings.Contains(string(body), "/fingerprint/script/") {
			return true
		}

	}

	return false
}

func isNewCaptchaChallenge(response *http.Response) bool {
	body, err := readAndCopyBody(response)
	if err != nil {
		return false
	}
	firstReg, err := regexp.MatchString(`cpo.src\s*=\s*[",']/cdn-cgi/challenge-platform/?\w?/?\w?/orchestrate/.*/v1`, string(body))
	if err != nil {
		return false
	}
	secondReg, err := regexp.MatchString(`window._cf_chl_opt`, string(body))
	if err != nil {
		return false
	}

	return strings.Contains(response.Header.Get("Server"), "cloudflare") &&
		(response.StatusCode == 403) &&
		firstReg && secondReg
}

func isNotRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var errno errNotRetryable
	if errors.As(err, &errno) {
		return true
	}

	return false
}

package hawk

import (
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type cfConfig struct {
	ApiDomain   string
	AuthParams  map[string]string
	Captcha     bool
	CaptchaFunc func(originalURL, siteKey string) (string, error)
	ErrorDelay  time.Duration
	FingerPrint bool
	MaxRetries  int
}

type cfChallengeState struct {
	baseObj                 string
	captchaResponseAPI      apiResponse
	challengePayloadBody    string
	domain                  string
	finalApi                *apiResponse
	finalResponse           *http.Response
	firstCaptchaResult      apiResponse
	initialResponseBody     []byte
	initURL                 string
	keyStrUriSafe           string
	mainPayloadResponseBody string
	md                      string
	name                    string
	originalResponse        *http.Response
	originalResponseBody    []byte
	p1Result                string
	p1TargetURL             string
	requestPass             string
	requestR                string
	requestURL              string
	result                  string
	startTime               time.Time
	ts                      int
	urlPart                 string
}

type apiResponse struct {
	BaseObj      string `json:"baseobj"`
	Captcha      bool   `json:"captcha"`
	CfChCpReturn string `json:"cf_ch_cp_return"`
	Click        bool   `json:"click"`
	JschlAnswer  string `json:"jschl_answer"`
	JschlVc      string `json:"jschl_vc"`
	Md           string `json:"md"`
	Name         string `json:"name"`
	Pass         string `json:"pass"`
	R            string `json:"r"`
	Result       string `json:"result"`
	ResultURL    string `json:"result_url"`
	SiteKey      string `json:"sitekey"`
	Status       string `json:"status"`
	TS           int    `json:"ts"`
	URL          string `json:"url"`
	Valid        bool   `json:"valid"`
}

type errNotRetryable struct {
	err error
}

func (e errNotRetryable) Error() string {
	return e.err.Error()
}

func (e errNotRetryable) Unwrap() error {
	return e.err
}

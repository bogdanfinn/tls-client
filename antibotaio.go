package tls_client

import (
	"io"

	http "github.com/bogdanfinn/fhttp"
)

type interstetial struct {
	RT     string `json:"rt"`
	CID    string `json:"cid"`
	HSH    string `json:"hsh"`
	T      string `json:"t"`
	S      int    `json:"s"`
	E      string `json:"e"`
	Host   string `json:"host"`
	IFS    string `json:"ifs"`
	Cookie string `json:"cookie"`
}

type slider struct {
	RT     string `json:"rt"`
	CID    string `json:"cid"`
	HSH    string `json:"hsh"`
	B      int    `json:"b"`
	S      int    `json:"s"`
	Host   string `json:"host"`
	IFS    string `json:"ifs"`
	Cookie string `json:"cookie"`
}

type antibotAIO struct {
	XAPIKey string
	Hosts   []string
	interstetial
	slider
}

func NewAntibotAIO(xAPIKey string, hosts []string) *antibotAIO {
	return &antibotAIO{
		XAPIKey: xAPIKey,
		Hosts:   hosts,
	}
}

func (c *httpClient) handleResponse(resp *http.Response) error {
	c.logger.Info("Handling response")

	// Nothing to solve, early return
	if resp.StatusCode == 200 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if checkForInterstetial(string(body)) {
		c.logger.Info("Interstetial detected... solving")
		interstetial, err := parseInterstetial(string(body))
		if err != nil {
			return err
		}

		c.config.antibotaio.interstetial = *interstetial
	}

	if checkForSliderCaptcha(string(body)) {
		c.logger.Info("Slider captcha detected... solving")
	}

	return nil
}

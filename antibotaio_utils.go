package tls_client

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// checkForInterstetial checks if the HTML body contains an interstetial challenge
func checkForInterstetial(body string) bool {
	return regexp.MustCompile(`(?i)<p id="cmsg">.*?enable\s*JS.*?ad\s*blocker.*?</p>.*?<script.*?>.*?var\s+dd\s*=\s*\{.*?'rt'\s*:\s*'.*?',.*?'host'\s*:\s*'[^']*captcha-delivery\.com'.*?\}.*?</script>`).MatchString(body)
}

func parseInterstetial(html string) (*interstetial, error) {
	// Regex pattern to match dd={...} pattern
	pattern := regexp.MustCompile(`var\s+dd=(\{[^}]+\})`)

	// Find the dd dictionary in the HTML
	matches := pattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil, fmt.Errorf("dd dictionary not found in HTML")
	}

	// Extract the JSON string
	jsonStr := matches[1]

	// Convert single quotes to double quotes
	// First, replace any escaped single quotes
	jsonStr = strings.ReplaceAll(jsonStr, `\'`, "'")

	// Then replace the outer single quotes with double quotes
	jsonStr = regexp.MustCompile(`'([^']+)'(\s*[:\,\}])`).ReplaceAllString(jsonStr, `"$1"$2`)

	// Create a DDDict struct to store the result
	result := &interstetial{}

	// Unmarshal the JSON string into the struct
	err := json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dd dictionary: %v", err)
	}

	return result, nil
}

func checkForSliderCaptcha(body string) bool {
	return regexp.MustCompile(`TODO`).MatchString(body)
}

func parseSliderCaptcha(body string) (*slider, error) {
	//TODO
	return nil, nil
}

func (a *antibotAIO) isHostnameInList(hostname string) bool {
	for _, host := range a.Hosts {
		if host == hostname {
			return true
		}
	}
	return false
}

package hawk

import "github.com/bogdanfinn/fhttp"

var initHeaders = http.Header{
	"sec-ch-ua":                 {`" Not;A Brand";v="99", "Google Chrome";v="91", "Chromium";v="91"`},
	"sec-ch-ua-mobile":          {"?0"},
	"upgrade-insecure-requests": {"1"},
	"user-agent":                {""},
	"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	"sec-fetch-site":            {"none"},
	"sec-fetch-mode":            {"navigate"},
	"sec-fetch-user":            {"?1"},
	"sec-fetch-dest":            {"document"},
	"accept-encoding":           {"gzip, deflate, br"},
	"accept-language":           {"en-US,en;q=0.9"},
	http.HeaderOrderKey:         {"sec-ch-ua", "sec-ch-ua-mobile", "upgrade-insecure-requests", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-site", "sec-fetch-dest", "accept-encoding", "accept-language"},
	http.PHeaderOrderKey:        {":method", ":authority", ":scheme", ":path"},
}

var challengeHeaders = http.Header{
	"user-agent":         {""},
	"cf-challenge":       {"b6245c8f8a8cb25"},
	"content-type":       {"application/x-www-form-urlencoded"},
	"accept":             {"*/*"},
	"origin":             {"https://www.origin.com"},
	"sec-fetch-site":     {"same-origin"},
	"sec-fetch-mode":     {"cors"},
	"sec-fetch-dest":     {"empty"},
	"referer":            {"https://www.referer.com/"},
	"accept-encoding":    {"gzip, deflate, br"},
	"accept-language":    {"en-US,en;q=0.9"},
	http.HeaderOrderKey:  {"content-length", "user-agent", "cf-challenge", "content-type", "accept", "origin", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language"},
	http.PHeaderOrderKey: {":method", ":authority", ":scheme", ":path"},
}

var submitHeaders = http.Header{
	"pragma":                    {"no-cache"},
	"cache-control":             {"max-age=0"},
	"upgrade-insecure-requests": {"1"},
	"origin":                    {"https://www.origin.com"},
	"content-type":              {"application/x-www-form-urlencoded"},
	"user-agent":                {""},
	"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
	"sec-fetch-site":            {"same-origin"},
	"sec-fetch-mode":            {"navigate"},
	"sec-fetch-dest":            {"document"},
	"referer":                   {"https://www.referer.com/"},
	"accept-encoding":           {"gzip, deflate, br"},
	"accept-language":           {"en-US,en;q=0.9"},
	http.HeaderOrderKey:         {"content-length", "pragma", "cache-control", "upgrade-insecure-requests", "origin", "content-type", "user-agent", "accept", "sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest", "referer", "accept-encoding", "accept-language"},
	http.PHeaderOrderKey:        {":method", ":authority", ":scheme", ":path"},
}

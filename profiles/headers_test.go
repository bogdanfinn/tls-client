package profiles

import (
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"
)

// What Chrome 152 on Windows sent to a local server for a top-level
// navigation, read off the HTTP/2 frames on 2026-09-08. Host and Connection
// come from the HTTP/1.1 reading of the same browser; the client writes Host
// itself, so it is in the order and not in the set.
var chrome152 = []struct{ name, value string }{
	{"host", ""},
	{"connection", "keep-alive"},
	{"sec-ch-ua", `"Chromium";v="152", "Not?A_Brand";v="24", "Google Chrome";v="152"`},
	{"sec-ch-ua-mobile", "?0"},
	{"sec-ch-ua-platform", `"Windows"`},
	{"upgrade-insecure-requests", "1"},
	{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/152.0.0.0 Safari/537.36"},
	{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	{"sec-fetch-site", "none"},
	{"sec-fetch-mode", "navigate"},
	{"sec-fetch-user", "?1"},
	{"sec-fetch-dest", "document"},
	{"accept-encoding", "gzip, deflate, br, zstd"},
	{"accept-language", "en-US,en;q=0.9"},
	{"priority", "u=0, i"},
}

func TestDefaultHeadersMatchWhatChrome152Sent(t *testing.T) {
	headers, ok := Chrome_152.DefaultHeaders()
	if !ok {
		t.Fatal("Chrome_152 has no default headers")
	}

	order := headers[http.HeaderOrderKey]
	if len(order) != len(chrome152) {
		t.Fatalf("the order names %d headers, Chrome sent %d:\n%v", len(order), len(chrome152), order)
	}
	for i, want := range chrome152 {
		if order[i] != want.name {
			t.Errorf("position %d: %s, Chrome sends %s", i, order[i], want.name)
		}
		if want.name == "host" {
			continue
		}
		got, found := valueOf(headers, want.name)
		if !found {
			t.Errorf("%s is missing", want.name)
		} else if got != want.value {
			t.Errorf("%s is %q, Chrome sends %q", want.name, got, want.value)
		}
	}
}

// valueOf finds a header whatever its spelling, since the set uses the
// browser's HTTP/1.1 spelling and the order is lower case.
func valueOf(h http.Header, name string) (string, bool) {
	for k, v := range h {
		if strings.EqualFold(k, name) && len(v) > 0 {
			return v[0], true
		}
	}
	return "", false
}

// The names are spelled as Chrome spells them over HTTP/1.1, where the
// spelling reaches the wire: the client hints in lower case, the rest with
// capitals.
func TestDefaultHeadersUseTheBrowsersSpelling(t *testing.T) {
	headers, _ := Chrome_152.DefaultHeaders()
	for _, name := range []string{"sec-ch-ua", "sec-ch-ua-mobile", "sec-ch-ua-platform", "User-Agent", "Accept", "Sec-Fetch-Site", "Upgrade-Insecure-Requests", "Accept-Encoding", "Accept-Language", "Connection"} {
		if _, ok := headers[name]; !ok {
			t.Errorf("no header spelled %q", name)
		}
	}
	brave, _ := Brave_146.DefaultHeaders()
	if _, ok := brave["Sec-GPC"]; !ok {
		t.Error("Brave's Sec-GPC is not spelled the way Brave spells it")
	}
}

// Five brand lists read off the wire. 148 is the one that tells a placement
// from a selection: reading the permutation table the other way round gives
// Chromium, Not/A)Brand, Brave for it.
func TestSecChUAMatchesTheCaptures(t *testing.T) {
	captures := []struct {
		major int
		brand string
		sent  string
	}{
		{148, "Brave", `"Chromium";v="148", "Brave";v="148", "Not/A)Brand";v="99"`},
		{151, "Brave", `"Not=A?Brand";v="99", "Brave";v="151", "Chromium";v="151"`},
		{152, "Brave", `"Chromium";v="152", "Not?A_Brand";v="24", "Brave";v="152"`},
		{152, "Google Chrome", `"Chromium";v="152", "Not?A_Brand";v="24", "Google Chrome";v="152"`},
		{152, "Microsoft Edge", `"Chromium";v="152", "Not?A_Brand";v="24", "Microsoft Edge";v="152"`},
		// The example in issue #213, which its author took from a real Chrome 133.
		{133, "Google Chrome", `"Not(A:Brand";v="99", "Google Chrome";v="133", "Chromium";v="133"`},
	}
	for _, c := range captures {
		if got := secChUA(c.major, c.brand); got != c.sent {
			t.Errorf("%s %d:\n got  %s\n sent %s", c.brand, c.major, got, c.sent)
		}
	}
}

func TestOnlyChromeAndBraveHaveDefaultHeaders(t *testing.T) {
	for name, profile := range map[string]ClientProfile{
		"firefox_133": Firefox_133, "firefox_148": Firefox_148, "safari_16_0": Safari_16_0,
		"safari_ios_18_0": Safari_IOS_18_0, "opera_90": Opera_90, "okhttp4_android_13": Okhttp4Android13,
		"zalando_ios_mobile": ZalandoIosMobile, "cloudscraper": CloudflareCustom,
	} {
		if headers, ok := profile.DefaultHeaders(); ok || headers != nil {
			t.Errorf("%s has default headers: %v", name, headers)
		}
	}
	for name, profile := range map[string]ClientProfile{"chrome_103": Chrome_103, "chrome_152_PSK": Chrome_152_PSK, "brave_146": Brave_146} {
		if _, ok := profile.DefaultHeaders(); !ok {
			t.Errorf("%s has no default headers", name)
		}
	}
}

func TestBraveDiffersFromChromeWhereBraveDiffers(t *testing.T) {
	headers, _ := Brave_146.DefaultHeaders()
	if got, _ := valueOf(headers, "sec-ch-ua"); got != `"Chromium";v="146", "Not-A.Brand";v="24", "Brave";v="146"` {
		t.Errorf("sec-ch-ua: %s", got)
	}
	if got, _ := valueOf(headers, "accept"); strings.Contains(got, "signed-exchange") {
		t.Errorf("Brave offers signed exchanges: %s", got)
	}
	if got, _ := valueOf(headers, "user-agent"); !strings.Contains(got, "Chrome/146.0.0.0") || strings.Contains(got, "Brave") {
		t.Errorf("Brave's User-Agent is Chrome's, unchanged: %s", got)
	}
	order := headers[http.HeaderOrderKey]
	for i, n := range order {
		if n == "sec-gpc" && (i == 0 || order[i-1] != "accept") {
			t.Errorf("sec-gpc follows %s, Brave sends it after accept", order[i-1])
		}
	}
}

func TestWhatFollowsTheMajor(t *testing.T) {
	tests := []struct {
		profile  ClientProfile
		zstd     bool
		priority bool
	}{
		{Chrome_120, false, false},
		{Chrome_124, true, true},
		{Chrome_133, true, true},
	}
	for _, tt := range tests {
		headers, _ := tt.profile.DefaultHeaders()
		encoding, _ := valueOf(headers, "accept-encoding")
		if strings.Contains(encoding, "zstd") != tt.zstd {
			t.Errorf("%s: accept-encoding is %q", tt.profile.GetClientHelloStr(), encoding)
		}
		if _, ok := valueOf(headers, "priority"); ok != tt.priority {
			t.Errorf("%s: priority present = %v", tt.profile.GetClientHelloStr(), ok)
		}
	}
}

// Every header is in the order and every name in the order is a header, host
// aside, for every mapped profile. A header left out of the order lands after
// the ordered ones, quietly.
func TestEveryDefaultHeaderIsOrdered(t *testing.T) {
	for name, profile := range MappedTLSClients {
		headers, ok := profile.DefaultHeaders()
		if !ok {
			continue
		}
		order := headers[http.HeaderOrderKey]
		seen := map[string]int{}
		for _, n := range order {
			seen[n]++
			if n != "host" {
				if _, found := valueOf(headers, n); !found {
					t.Errorf("%s: %s is in the order and not in the set", name, n)
				}
			}
		}
		for k := range headers {
			if k == http.HeaderOrderKey {
				continue
			}
			if seen[strings.ToLower(k)] != 1 {
				t.Errorf("%s: %s appears %d times in the order", name, k, seen[strings.ToLower(k)])
			}
		}
	}
}

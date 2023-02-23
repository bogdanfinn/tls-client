package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	goHttp "net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	"github.com/bogdanfinn/fhttp/httptrace"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/shared"
	tls "github.com/bogdanfinn/utls"
	"github.com/google/uuid"
)

func main() {
	sslPinning()
	requestToppsAsGoClient()
	requestToppsAsChrome107Client()
	postAsTlsClient()
	requestWithFollowRedirectSwitch()
	requestWithCustomClient()
	http2ReuseTlsClient()
	//rotateProxiesOnClient() //commented out because no proxies committed
	http2HeaderFrameOrder()
	loginZalandoMobileAndroid()
	downloadImageWithTlsClient()
}

func sslPinning() {
	jar := tls_client.NewCookieJar()

	//	I generated the pins by running the following command:
	//	➜ hpkp-pins -server=bstn.com:443

	pins := map[string][]string{
		"bstn.com": {
			"NQvy9sFS99nBqk/nZCUF44hFhshrkvxqYtfrZq3i+Ww=",
			"4a6cPehI7OG6cuDZka5NDZ7FR8a60d3auda+sKfg4Ng=",
			"x4QzPSC810K5/cMjb05Qm4k3Bw5zBn4lTdO/nEW/Td4=",
		},
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(60),
		tls_client.WithClientProfile(tls_client.Chrome_108),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithCookieJar(jar),
		tls_client.WithCertificatePinning(pins, tls_client.DefaultBadPinHandler),
		tls_client.WithCharlesProxy("127.0.0.1", "8888"),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	u := "https://bstn.com"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":             {"*/*"},
		"accept-encoding":    {"gzip, deflate, br"},
		"accept-language":    {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"sec-ch-ua":          {`"Google Chrome";v="107", "Chromium";v="107", "Not=A?Brand";v="24"`},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {`"macOS"`},
		"sec-fetch-dest":     {"empty"},
		"user-agent":         {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"user-agent",
		},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	resp.Body.Close()

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("GET %s : %d\n", u, resp.StatusCode)
}

func requestToppsAsGoClient() {
	c := &goHttp.Client{}

	r, err := goHttp.NewRequest(http.MethodGet, "https://www.topps.com/", nil)
	if err != nil {
		log.Println(err)
		return
	}

	r.Header = goHttp.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
	}

	requestBytes, err := httputil.DumpRequestOut(r, r.ContentLength > 0)

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("raw request bytes sent over wire: %d (%d kb)\n", len(requestBytes), len(requestBytes)/1024)

	re, err := c.Do(r)

	if err != nil {
		log.Println(err)
		return
	}

	defer re.Body.Close()

	responseBytes, err := httputil.DumpResponse(re, re.ContentLength > 0)

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("raw response bytes received over wire: %d (%d kb)\n", len(responseBytes), len(responseBytes)/1024)

	log.Printf("requesting topps as golang => status code: %d\n", re.StatusCode)
}

func requestToppsAsChrome107Client() {
	jar := tls_client.NewCookieJar()

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_107),
		tls_client.WithDebug(),
		//tls_client.WithProxyUrl("http://user:pass@host:port"),
		//tls_client.WithNotFollowRedirects(),
		//tls_client.WithInsecureSkipVerify(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.topps.com/", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting topps as chrome107 => status code: %d\n", resp.StatusCode)

	u, err := url.Parse("https://www.topps.com/")

	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("tls client cookies for url %s : %v\n", u.String(), client.GetCookies(u))
}

func postAsTlsClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_107),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	postData := url.Values{}
	postData.Add("foo", "bar")
	postData.Add("baz", "foo")

	req, err := http.NewRequest(http.MethodPost, "https://eonk4gg5hquk0g6.m.pipedream.net", strings.NewReader(postData.Encode()))
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":          {"*/*"},
		"content-type":    {"application/x-www-form-urlencoded"},
		"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"user-agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"content-type",
			"accept-language",
			"user-agent",
			"content-length",
			"host",
		},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("POST Request status code: %d\n", resp.StatusCode)
}

func requestWithFollowRedirectSwitch() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_107),
		tls_client.WithNotFollowRedirects(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://currys.co.uk/products/sony-playstation-5-digital-edition-825-gb-10205198.html", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"Accept-Encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting currys.co.uk without automatic redirect follow => status code: %d (Redirect Not Folloed)\n", resp.StatusCode)

	client.SetFollowRedirect(true)

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting currys.co.uk with automatic redirect follow => status code: %d (Redirect Followed)\n", resp.StatusCode)
}

func downloadImageWithTlsClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_107),
		tls_client.WithNotFollowRedirects(),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://avatars.githubusercontent.com/u/17678241?v=4", nil)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	log.Printf("requesting image => status code: %d\n", resp.StatusCode)

	ex, err := os.Executable()

	if err != nil {
		log.Println(err)
		return
	}

	exPath := filepath.Dir(ex)

	fileName := fmt.Sprintf("%s/%s", exPath, "example-test.jpg")

	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("wrote file to: %s\n", fileName)
}

func http2HeaderFrameOrder() {
	firefoxOptions := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Firefox_106),
	}

	chromeOptions := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_108),
	}

	firefoxClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), firefoxOptions...)
	if err != nil {
		log.Println(err)
		return
	}

	chromeClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), chromeOptions...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
	if err != nil {
		log.Println(err)
		return
	}

	firefoxResp, err := firefoxClient.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer firefoxResp.Body.Close()

	chromeResp, err := chromeClient.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer chromeResp.Body.Close()

	firefoxReadBytes, err := io.ReadAll(firefoxResp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse := shared.TlsApiResponse{}
	if err := json.Unmarshal(firefoxReadBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	for i, frame := range tlsApiResponse.HTTP2.SentFrames {
		log.Printf("Firefox Frame %d: %s: %d\n", i, frame.FrameType, frame.StreamID)

		if frame.FrameType == "HEADERS" {
			log.Printf("Firefox Header Priority: %v\n", frame.Priority)
		}
	}

	chromeReadBytes, err := io.ReadAll(chromeResp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse = shared.TlsApiResponse{}
	if err := json.Unmarshal(chromeReadBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	for i, frame := range tlsApiResponse.HTTP2.SentFrames {
		log.Printf("Chrome Frame %d: %s: %d\n", i, frame.FrameType, frame.StreamID)

		if frame.FrameType == "HEADERS" {
			log.Printf("Chrome Header Priority: %v\n", frame.Priority)
		}
	}
}

// func rotateProxiesOnClient() {
// 	options := []tls_client.HttpClientOption{
// 		tls_client.WithTimeoutSeconds(30),
// 		tls_client.WithClientProfile(tls_client.Chrome_107),
// 		tls_client.WithProxyUrl("http://user:pass@host:port"),
// 	}

// 	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	req.Header = http.Header{
// 		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
// 		"accept-encoding":           {"gzip"},
// 		"Accept-Encoding":           {"gzip"},
// 		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
// 		"cache-control":             {"max-age=0"},
// 		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
// 		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
// 		"sec-ch-ua-mobile":          {"?0"},
// 		"sec-ch-ua-platform":        {`"macOS"`},
// 		"sec-fetch-dest":            {"document"},
// 		"sec-fetch-mode":            {"navigate"},
// 		"sec-fetch-site":            {"none"},
// 		"sec-fetch-user":            {"?1"},
// 		"upgrade-insecure-requests": {"1"},
// 		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
// 		http.HeaderOrderKey: {
// 			"accept",
// 			"accept-encoding",
// 			"accept-language",
// 			"cache-control",
// 			"if-none-match",
// 			"sec-ch-ua",
// 			"sec-ch-ua-mobile",
// 			"sec-ch-ua-platform",
// 			"sec-fetch-dest",
// 			"sec-fetch-mode",
// 			"sec-fetch-site",
// 			"sec-fetch-user",
// 			"upgrade-insecure-requests",
// 			"user-agent",
// 		},
// 	}

// 	resp, err := client.Do(req)

// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	defer resp.Body.Close()

// 	readBytes, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	tlsApiResponse := shared.TlsApiResponse{}
// 	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	log.Println(fmt.Sprintf("requesting tls.peet.ws with proxy 1 => ip: %s", tlsApiResponse.IP))

// 	// you need to put in here a valid proxy to make the example work
// 	err = client.SetProxy("http://user:pass@host:port")
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	resp, err = client.Do(req)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	defer resp.Body.Close()

// 	readBytes, err = io.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	tlsApiResponse = shared.TlsApiResponse{}
// 	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	log.Println(fmt.Sprintf("requesting tls.peet.ws with proxy 2 => ip: %s", tlsApiResponse.IP))
// }

func requestWithCustomClient() {
	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      65536,
		http2.SettingMaxConcurrentStreams: 1000,
		http2.SettingInitialWindowSize:    6291456,
		http2.SettingMaxHeaderListSize:    262144,
	}
	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxHeaderListSize,
	}

	pseudoHeaderOrder := []string{
		":method",
		":authority",
		":scheme",
		":path",
	}

	connectionFlow := uint32(15663105)

	specFactory := func() (tls.ClientHelloSpec, error) {
		return tls.ClientHelloSpec{
			CipherSuites: []uint16{
				tls.GREASE_PLACEHOLDER,
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			CompressionMethods: []uint8{
				tls.CompressionNone,
			},
			Extensions: []tls.TLSExtension{
				&tls.UtlsGREASEExtension{},
				&tls.SNIExtension{},
				&tls.UtlsExtendedMasterSecretExtension{},
				&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
				&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
					tls.CurveID(tls.GREASE_PLACEHOLDER),
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
				}},
				&tls.SupportedPointsExtension{SupportedPoints: []byte{
					0,
				}},
				&tls.SessionTicketExtension{},
				&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&tls.StatusRequestExtension{},
				&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
					tls.ECDSAWithP256AndSHA256,
					tls.PSSWithSHA256,
					tls.PKCS1WithSHA256,
					tls.ECDSAWithP384AndSHA384,
					tls.PSSWithSHA384,
					tls.PKCS1WithSHA384,
					tls.PSSWithSHA512,
					tls.PKCS1WithSHA512,
				}},
				&tls.SCTExtension{},
				&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
					{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
					{Group: tls.X25519},
				}},
				&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
					tls.PskModeDHE,
				}},
				&tls.SupportedVersionsExtension{Versions: []uint16{
					tls.VersionTLS13,
					tls.VersionTLS12,
					tls.VersionTLS11,
					tls.VersionTLS10,
				}},
				&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{
					tls.CertCompressionBrotli,
				}},
				&tls.ApplicationSettingsExtension{SupportedProtocols: []string{"h2"}},
				&tls.UtlsGREASEExtension{},
				&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
			},
		}, nil
	}

	customClientProfile := tls_client.NewClientProfile(tls.ClientHelloID{
		Client:      "MyCustomProfile",
		Version:     "1",
		Seed:        nil,
		SpecFactory: specFactory,
	}, settings, settingsOrder, pseudoHeaderOrder, connectionFlow, nil, nil)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(60),
		tls_client.WithClientProfile(customClientProfile), // use custom profile here
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://www.topps.com/", nil)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"accept-encoding":           {"gzip"},
		"accept-language":           {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
		"cache-control":             {"max-age=0"},
		"if-none-match":             {`W/"4d0b1-K9LHIpKrZsvKsqNBKd13iwXkWxQ"`},
		"sec-ch-ua":                 {`"Google Chrome";v="105", "Not)A;Brand";v="8", "Chromium";v="105"`},
		"sec-ch-ua-mobile":          {"?0"},
		"sec-ch-ua-platform":        {`"macOS"`},
		"sec-fetch-dest":            {"document"},
		"sec-fetch-mode":            {"navigate"},
		"sec-fetch-site":            {"none"},
		"sec-fetch-user":            {"?1"},
		"upgrade-insecure-requests": {"1"},
		"user-agent":                {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: {
			"accept",
			"accept-encoding",
			"accept-language",
			"cache-control",
			"if-none-match",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting topps as customClient1 => status code: %d\n", resp.StatusCode)
}

type ZalandoLoginPayload struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	AppVersion     string `json:"appVersion"`
	AppdomainId    string `json:"appdomainId"`
	DeviceLanguage string `json:"deviceLanguage"`
	DevicePlatform string `json:"devicePlatform"`
	Sig            string `json:"sig"`
	Ts             int    `json:"ts"`
	Uuid           string `json:"uuid"`
}

func loginZalandoMobileAndroid() {
	// next to the uuid you need ts and sig and of course akamai sensor data
	id := uuid.New()
	akamaiBmpSensor := ""
	ts := 1661985341830
	sig := "f01ae091f136195da14333dc7485e0099dd8fb3a"

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(60),
		tls_client.WithClientProfile(tls_client.ZalandoAndroidMobile),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	// ts and sig has to match with ts and sig from headers
	loginPayload := ZalandoLoginPayload{
		Email:          "random@gmail.com",
		Password:       "randompassword",
		AppVersion:     "22.10.3",
		AppdomainId:    "1",
		DeviceLanguage: "en",
		DevicePlatform: "android",
		Sig:            sig,
		Ts:             ts,
		Uuid:           id.String(),
	}

	jsonLoginPayload, err := json.Marshal(loginPayload)
	if err != nil {
		log.Println(err)
		return
	}

	bodyBuffer := bytes.NewBuffer(jsonLoginPayload)
	req, err := http.NewRequest(http.MethodPost, "https://en.zalando.de/api/mobile/v3/user/login.json", bodyBuffer)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header = http.Header{
		"cache-control":        {"private, no-cache, no-store"},
		"x-app-domain":         {"1"},
		"user-agent":           {`Zalando/22.11.0 (Linux; Android 8.0.0; Samsung SM-A520F/R16NW.A520FXXUGCTKA)`},
		"x-uuid":               {id.String()},
		"x-ts":                 {strconv.Itoa(ts)},
		"x-device-language":    {"en"},
		"x-sig":                {sig},
		"x-os-version":         {"9"},
		"accept-language":      {"en-GB"},
		"accept":               {"application/json"},
		"x-app-version":        {"22.10.3"},
		"x-device-platform":    {"android"},
		"x-device-os":          {"android"},
		"x-zalando-mobile-app": {"1166c0792788b3f3a"},
		"x-logged-in":          {"false"},
		"x-advertising-id":     {"6fdbd95c-ccf1-40cf-9910-88f26deaa61f"},
		"content-type":         {"application/json"},
		"content-length":       {strconv.Itoa(bodyBuffer.Len())},
		"accept-encoding":      {"gzip"},
		"ot-tracer-traceid":    {"c71c9283de42cad1"},
		"ot-tracer-spanid":     {"b603dda8154a3f50"},
		"ot-tracer-sampled":    {"true"},
		"x-acf-sensor-data":    {akamaiBmpSensor},
		http.HeaderOrderKey: {
			"cache-control",
			"x-app-domain",
			"user-agent",
			"x-uuid",
			"x-ts",
			"x-device-language",
			"x-sig",
			"x-os-version",
			"accept-language",
			"accept",
			"x-app-version",
			"x-device-platform",
			"x-device-os",
			"x-zalando-mobile-app",
			"x-logged-in",
			"x-advertising-id",
			"content-type",
			"content-length",
			"accept-encoding",
			"ot-tracer-traceid",
			"ot-tracer-spanid",
			"ot-tracer-sampled",
			"x-acf-sensor-data",
		},
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Printf("requesting zalando login as zalando android client => status code: %d\n", resp.StatusCode)

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(readBytes))
}

func http2ReuseTlsClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_108),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	callInLoop := func(wg *sync.WaitGroup, id int, client tls_client.HttpClient, amount int, url string) {
		defer wg.Done()
		for i := 0; i < amount; i++ {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				log.Println(err)
				return
			}

			req.Header = http.Header{
				"accept":          {"*/*"},
				"accept-language": {"de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7"},
				"user-agent":      {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"},
				http.HeaderOrderKey: {
					"accept",
					"accept-language",
					"user-agent",
				},
			}

			clientTrace := &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) {
					log.Printf("Connection was reused in routine %d: %t", id, info.Reused)
				},
			}

			req = req.WithContext(httptrace.WithClientTrace(req.Context(), clientTrace))

			resp, err := client.Do(req)

			if err != nil {
				log.Println(err)
				return
			}

			if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
				log.Fatal(err)
				return
			}

			resp.Body.Close()

			log.Println(fmt.Sprintf("Go Routine %d: %s: status code: %d", id, url, resp.StatusCode))

			time.Sleep(2 * time.Second)
		}
	}

	log.Println("starting go routines to https://www.google.de/")
	var wg sync.WaitGroup
	wg.Add(2)

	go callInLoop(&wg, 1, client, 3, "https://www.google.de/")
	go callInLoop(&wg, 2, client, 3, "https://www.google.de/")

	wg.Wait()
}

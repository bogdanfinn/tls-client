package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/shared"
	tls "github.com/bogdanfinn/utls"
)

func main() {
	sslPinning()
	requestToppsAsChrome107Client()
	postAsTlsClient()
	requestWithFollowRedirectSwitch()
	requestWithCustomClient()
	rotateProxiesOnClient()
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

func requestToppsAsChrome107Client() {
	jar := tls_client.NewCookieJar()

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_107),
		tls_client.WithDebug(),
		tls_client.WithTransportOptions(&tls_client.TransportOptions{
			DisableKeepAlives:      false,
			DisableCompression:     false,
			MaxIdleConns:           0,
			MaxIdleConnsPerHost:    0,
			MaxConnsPerHost:        0,
			MaxResponseHeaderBytes: 0,
			WriteBufferSize:        0,
			ReadBufferSize:         0,
		}),
		tls_client.WithCookieJar(jar),
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

func rotateProxiesOnClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithClientProfile(tls_client.Chrome_107),
		tls_client.WithProxyUrl("http://user:pass@host:port"),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		log.Println(err)
		return
	}

	req, err := http.NewRequest(http.MethodGet, "https://tls.peet.ws/api/all", nil)
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

	readBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse := shared.TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	log.Println(fmt.Sprintf("requesting tls.peet.ws with proxy 1 => ip: %s", tlsApiResponse.IP))

	// you need to put in here a valid proxy to make the example work
	err = client.SetProxy("http://user:pass@host:port")
	if err != nil {
		log.Println(err)
		return
	}

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	readBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	tlsApiResponse = shared.TlsApiResponse{}
	if err := json.Unmarshal(readBytes, &tlsApiResponse); err != nil {
		log.Println(err)
		return
	}

	log.Println(fmt.Sprintf("requesting tls.peet.ws with proxy 2 => ip: %s", tlsApiResponse.IP))
}

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

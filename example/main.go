package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	goHttp "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/fhttp/cookiejar"
	"github.com/bogdanfinn/fhttp/http2"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/shared"
	tls "github.com/bogdanfinn/utls"
	"github.com/google/uuid"
)

func main() {
	shareHttpClientInGoRoutines()
	requestToppsAsGoClient()
	requestToppsAsChrome105Client()
	postAsTlsClient()
	requestWithFollowRedirectSwitch()
	requestWithCustomClient()
	rotateProxiesOnClient()
	downloadImageWithTlsClient()
	loginZalandoMobile()
}

func requestToppsAsGoClient() {
	c := &goHttp.Client{}

	r, err := goHttp.NewRequest(http.MethodGet, "https://www.topps.com/", nil)

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

	re, err := c.Do(r)

	if err != nil {
		log.Println(err)
		return
	}

	defer re.Body.Close()

	log.Println(fmt.Sprintf("requesting topps as golang => status code: %d", re.StatusCode))
}

func requestToppsAsChrome105Client() {
	cJar, _ := cookiejar.New(nil)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
		//tls_client.WithProxyUrl("http://user:pass@host:ip"),
		//tls_client.WithNotFollowRedirects(),
		//tls_client.WithInsecureSkipVerify(),
		tls_client.WithCookieJar(cJar), // create cookieJar instance and pass it as argument
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

	log.Println(fmt.Sprintf("requesting topps as chrome105 => status code: %d", resp.StatusCode))

	u, err := url.Parse("https://www.topps.com/")

	if err != nil {
		log.Println(err)
		return
	}

	log.Println(fmt.Sprintf("tls client cookies for url %s : %v", u.String(), client.GetCookies(u)))
}

func postAsTlsClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
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
		"content-type":    {"application/"},
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

	log.Println(fmt.Sprintf("POST Request status code: %d", resp.StatusCode))
}

func shareHttpClientInGoRoutines() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
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

			resp, err := client.Do(req)

			if err != nil {
				log.Println(err)
				return
			}

			defer resp.Body.Close()

			log.Println(fmt.Sprintf("Go Routine %d: %s: status code: %d", id, url, resp.StatusCode))

			time.Sleep(2 * time.Second)
		}
	}

	log.Println("starting go routines to https://example.com/")
	var wg sync.WaitGroup
	wg.Add(3)

	go callInLoop(&wg, 1, client, 5, "https://example.com/")
	go callInLoop(&wg, 2, client, 5, "https://example.com/")
	go callInLoop(&wg, 3, client, 5, "https://example.com/")

	wg.Wait()
}

func requestWithFollowRedirectSwitch() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
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

	log.Println(fmt.Sprintf("requesting currys.co.uk without automatic redirect follow => status code: %d (Redirect Not Folloed)", resp.StatusCode))

	client.SetFollowRedirect(true)

	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()

	log.Println(fmt.Sprintf("requesting currys.co.uk with automatic redirect follow => status code: %d (Redirect Followed)", resp.StatusCode))
}

func downloadImageWithTlsClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
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

	log.Println(fmt.Sprintf("requesting image => status code: %d", resp.StatusCode))

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

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(fmt.Sprintf("wrote file to: %s", fileName))
}

func rotateProxiesOnClient() {
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(30),
		tls_client.WithClientProfile(tls_client.Chrome_105),
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

	readBytes, err := ioutil.ReadAll(resp.Body)
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

	readBytes, err = ioutil.ReadAll(resp.Body)
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
				&tls.SupportedCurvesExtension{[]tls.CurveID{
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
				&tls.KeyShareExtension{[]tls.KeyShare{
					{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
					{Group: tls.X25519},
				}},
				&tls.PSKKeyExchangeModesExtension{[]uint8{
					tls.PskModeDHE,
				}},
				&tls.SupportedVersionsExtension{[]uint16{
					tls.VersionTLS13,
					tls.VersionTLS12,
					tls.VersionTLS11,
					tls.VersionTLS10,
				}},
				&tls.UtlsCompressCertExtension{[]tls.CertCompressionAlgo{
					tls.CertCompressionBrotli,
				}},
				&tls.ALPSExtension{SupportedProtocols: []string{"h2"}},
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
	}, settings, settingsOrder, pseudoHeaderOrder, connectionFlow, nil)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(60),
		tls_client.WithClientProfile(customClientProfile), // use custom profile here
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

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

	log.Println(fmt.Sprintf("requesting topps as customClient1 => status code: %d", resp.StatusCode))
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

func loginZalandoMobile() {
	// next to the uuid you need ts and sig and of course akamai sensor data
	id := uuid.New()
	akamaiBmpSensor := ""
	ts := 1661985341830
	sig := "f01ae091f136195da14333dc7485e0099dd8fb3a"

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(60),
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

	log.Println(fmt.Sprintf("requesting zalando login as zalando android client => status code: %d", resp.StatusCode))

	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(readBytes))
}

package tests

import (
	"github.com/bogdanfinn/tls-client/profiles"
	tls "github.com/bogdanfinn/utls"
)

type TlsApiResponse struct {
	IP          string `json:"ip"`
	HTTPVersion string `json:"http_version"`
	Method      string `json:"method"`
	TLS         struct {
		TLSVersionRecord     string   `json:"tls_version_record"`
		TLSVersionNegotiated string   `json:"tls_version_negotiated"`
		Ja3                  string   `json:"ja3"`
		Ja3Hash              string   `json:"ja3_hash"`
		ClientRandom         string   `json:"client_random"`
		SessionID            string   `json:"session_id"`
		Ciphers              []string `json:"ciphers"`
		Extensions           []struct {
			EllipticCurvesPointFormats interface{} `json:"elliptic_curves_point_formats,omitempty"`
			Name                       string      `json:"name"`
			ServerName                 string      `json:"server_name,omitempty"`
			Data                       string      `json:"data,omitempty"`
			PskKeyExchangeMode         string      `json:"PSK_Key_Exchange_Mode,omitempty"`
			SupportedGroups            []string    `json:"supported_groups,omitempty"`
			Protocols                  []string    `json:"protocols,omitempty"`
			SignatureAlgorithms        []string    `json:"signature_algorithms,omitempty"`
			SharedKeys                 []struct {
				TLSGrease0X7A7A string `json:"TLS_GREASE (0x7a7a),omitempty"`
				X2551929        string `json:"X25519 (29),omitempty"`
			} `json:"shared_keys,omitempty"`
			Versions      []string `json:"versions,omitempty"`
			Algorithms    []string `json:"algorithms,omitempty"`
			StatusRequest struct {
				CertificateStatusType   string `json:"certificate_status_type"`
				ResponderIDListLength   int    `json:"responder_id_list_length"`
				RequestExtensionsLength int    `json:"request_extensions_length"`
			} `json:"status_request,omitempty"`
			PaddingDataLength int `json:"padding_data_length,omitempty"`
		} `json:"extensions"`
	} `json:"tls"`
	HTTP2 struct {
		AkamaiFingerprint     string `json:"akamai_fingerprint"`
		AkamaiFingerprintHash string `json:"akamai_fingerprint_hash"`
		SentFrames            []struct {
			FrameType string   `json:"frame_type"`
			Settings  []string `json:"settings,omitempty"`
			Headers   []string `json:"headers,omitempty"`
			Flags     []string `json:"flags,omitempty"`
			Priority  struct {
				Weight    int `json:"weight"`
				DependsOn int `json:"depends_on"`
				Exclusive int `json:"exclusive"`
			} `json:"priority,omitempty"`
			Length    int `json:"length"`
			Increment int `json:"increment,omitempty"`
			StreamID  int `json:"stream_id,omitempty"`
		} `json:"sent_frames"`
	} `json:"http2"`
	HTTP1 struct {
		Headers []string `json:"headers"`
	} `json:"http1"`
}

const (
	chrome        = "chrome"
	firefox       = "firefox"
	opera         = "opera"
	safari        = "safari"
	safariIpadOs  = "safari_Ipad"
	safariIos     = "safari_IOS"
	okhttpAndroid = "okhttp_Android"

	peetApiEndpoint = "https://tls.peet.ws/api/all"

	ja3String             = "ja3String"
	ja3Hash               = "ja3Hash"
	akamaiFingerprint     = "akamaiFingerprint"
	akamaiFingerprintHash = "akamaiFingerprintHash"
)

var clientFingerprints = map[string]map[string]map[string]string{
	chrome: {
		profiles.Chrome_133.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,35-13-17613-51-18-11-43-5-16-0-65037-27-10-45-23-65281,4588-29-23-24,0",
			ja3Hash:               "74e530e488a43fddd78be75918be78c7",
			akamaiFingerprint:     "1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "52d84b11737d980aef856699f885ca86",
		},
		profiles.Chrome_131.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,13-65037-65281-18-27-16-5-10-17513-11-51-35-43-45-0-23,4588-29-23-24,0",
			ja3Hash:               "a19ab9f02aacf42deddc1f2acb3d3f63",
			akamaiFingerprint:     "1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "52d84b11737d980aef856699f885ca86",
		},
		profiles.Chrome_124.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,27-18-23-17513-16-43-13-11-0-35-10-65037-5-65281-45-51,25497-29-23-24,0",
			ja3Hash:               "64aff24dbef210f33880d4f62e1493dd",
			akamaiFingerprint:     "1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "52d84b11737d980aef856699f885ca86",
		},
		profiles.Chrome_120.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-45-43-5-23-35-13-65281-16-65037-18-51-10-11-17513-27,29-23-24,0",
			ja3Hash:               "1d9a054bac1eef41f30d370f9bbb2ad2",
			akamaiFingerprint:     "1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "52d84b11737d980aef856699f885ca86",
		},
		profiles.Chrome_117.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,45-0-16-13-43-17513-10-23-35-27-18-5-51-65281-11-21,29-23-24,0",
			ja3Hash:               "1ddf8a0ebd957d10c1ab320b10450028",
			akamaiFingerprint:     "1:65536;2:0;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "52d84b11737d980aef856699f885ca86",
		},
		tls.HelloChrome_112_PSK.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,45-51-17513-43-0-11-5-23-16-10-65281-27-18-35-13-21-41,29-23-24,0",
			ja3Hash:               "11d372983aac706304b678a44351c8dd",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_112.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,45-51-17513-43-0-11-5-23-16-10-65281-27-18-35-13-21,29-23-24,0",
			ja3Hash:               "7f052aeccc9b50e9b3a43a02780539b2",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_111.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,27-11-17513-5-10-18-23-0-45-51-43-35-65281-16-13-21,29-23-24,0",
			ja3Hash:               "499d7c2439dc2fb83d1ab2e52b9dc680",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_110.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,23-27-18-51-17513-0-16-35-11-5-65281-43-13-45-10-21,29-23-24,0",
			ja3Hash:               "f30e7d05622c38802b2ee65d147f4df8",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_109.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_108.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_107.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_106.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;2:0;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "a345a694846ad9f6c97bcc3c75adbe26",
		},
		tls.HelloChrome_105.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "4f04edce68a7ecbe689edce7bf5f23f3",
		},
		tls.HelloChrome_104.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "4f04edce68a7ecbe689edce7bf5f23f3",
		},
		tls.HelloChrome_103.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "4f04edce68a7ecbe689edce7bf5f23f3",
		},
	},
	firefox: {
		tls.HelloFirefox_102.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		tls.HelloFirefox_104.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		tls.HelloFirefox_105.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		tls.HelloFirefox_106.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		tls.HelloFirefox_108.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		tls.HelloFirefox_110.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-16-5-34-51-43-13-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "ad55557b7cbd735c2627f7ebb3b3d493",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		profiles.Firefox_117.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536;4:131072;5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "3d9132023bf26a71d40fe766e5c24c9d",
		},
		profiles.Firefox_132.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-16-5-34-51-43-13-28-27-65037,4588-29-23-24-25-256-257,0",
			ja3Hash:               "a767f8ae9115cc5752e5cff59612e74f",
			akamaiFingerprint:     "1:65536;2:0;4:131072;5:16384;9:1|12517377|0|m,p,a,s",
			akamaiFingerprintHash: "a80d4d15d0c3bdd7b34b39d61cdaf0f7",
		},
	},
	opera: {
		tls.HelloOpera_89.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "4f04edce68a7ecbe689edce7bf5f23f3",
		},
		tls.HelloOpera_90.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "4f04edce68a7ecbe689edce7bf5f23f3",
		},
		tls.HelloOpera_91.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536;3:1000;4:6291456;6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "4f04edce68a7ecbe689edce7bf5f23f3",
		},
	},
	safari: {
		tls.HelloSafari_15_6_1.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:4194304;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "dda308d35f4e5db7b52a61720ca1b122",
		},
		tls.HelloSafari_16_0.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:4194304;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "dda308d35f4e5db7b52a61720ca1b122",
		},
	},
	safariIpadOs: {
		tls.HelloIPad_15_6.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "d5fcbdc393757341115a861bf8d23265",
		},
	},
	safariIos: {
		tls.HelloIOS_15_5.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "d5fcbdc393757341115a861bf8d23265",
		},
		tls.HelloIOS_15_6.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "d5fcbdc393757341115a861bf8d23265",
		},
		tls.HelloIOS_16_0.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "d5fcbdc393757341115a861bf8d23265",
		},
		profiles.Safari_IOS_17_0.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "2:0;4:2097152;3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "ad8424af1cc590e09f7b0c499bf7fcdb",
		},
		profiles.Safari_IOS_18_0.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "2:0;3:100;4:2097152;:1;9:1|10420225|0|m,s,a,p",
			akamaiFingerprintHash: "62317f06028f316631c157c720223e43",
		},
	},
	okhttpAndroid: {
		profiles.Okhttp4Android13.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			ja3Hash:               "f79b6bad2ad0641e1921aef10262856b",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
		profiles.Okhttp4Android12.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			ja3Hash:               "f79b6bad2ad0641e1921aef10262856b",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
		profiles.Okhttp4Android11.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			ja3Hash:               "f79b6bad2ad0641e1921aef10262856b",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
		profiles.Okhttp4Android10.GetClientHelloStr(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-51-45-43-21,29-23-24,0",
			ja3Hash:               "f79b6bad2ad0641e1921aef10262856b",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
		profiles.Okhttp4Android9.GetClientHelloStr(): map[string]string{
			ja3String:             "771,49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0",
			ja3Hash:               "6f5e62edfa5933b1332ddf8b9fb3ef9d",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
		profiles.Okhttp4Android8.GetClientHelloStr(): map[string]string{
			ja3String:             "771,49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,65281-0-23-35-13-5-16-11-10,29-23-24,0",
			ja3Hash:               "6f5e62edfa5933b1332ddf8b9fb3ef9d",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
		profiles.Okhttp4Android7.GetClientHelloStr(): map[string]string{
			ja3String:             "771,49195-49196-52393-49199-49200-52392-49171-49172-156-157-47-53,65281-0-23-35-13-16-11-10,23-24-25,0",
			ja3Hash:               "f6a0bfafe2bf7d9c79ffb3f269b64b46",
			akamaiFingerprint:     "4:16777216|16711681|0|m,p,a,s",
			akamaiFingerprintHash: "605a1154008045d7e3cb3c6fb062c0ce",
		},
	},
}

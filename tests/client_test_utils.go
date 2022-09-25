package tests

import tls "github.com/bogdanfinn/utls"

const (
	chrome       = "chrome"
	firefox      = "firefox"
	opera        = "opera"
	safari       = "safari"
	safariIpadOs = "safari_Ipad"
	safariIos    = "safari_IOS"

	peetApiEndpoint = "https://tls.peet.ws/api/all"
	//peetApiEndpoint = "https://tls.cloudscraper.io/api/all"

	ja3String             = "ja3String"
	ja3StringWithPadding  = "ja3StringPadding"
	ja3Hash               = "ja3Hash"
	ja3HashWithPadding    = "ja3HashPadding"
	akamaiFingerprint     = "akamaiFingerprint"
	akamaiFingerprintHash = "akamaiFingerprintHash"
)

type TlsApiResponse struct {
	Donate      string `json:"donate"`
	IP          string `json:"ip"`
	HTTPVersion string `json:"http_version"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	TLS         struct {
		Ciphers    []string `json:"ciphers"`
		Extensions []struct {
			Name                       string      `json:"name"`
			ServerName                 string      `json:"server_name,omitempty"`
			Data                       string      `json:"data,omitempty"`
			SupportedGroups            []string    `json:"supported_groups,omitempty"`
			EllipticCurvesPointFormats interface{} `json:"elliptic_curves_point_formats,omitempty"`
			Protocols                  []string    `json:"protocols,omitempty"`
			StatusRequest              struct {
				CertificateStatusType   string `json:"certificate_status_type"`
				ResponderIDListLength   int    `json:"responder_id_list_length"`
				RequestExtensionsLength int    `json:"request_extensions_length"`
			} `json:"status_request,omitempty"`
			SignatureAlgorithms []string `json:"signature_algorithms,omitempty"`
			SharedKeys          []struct {
				TLSGrease0Xfafa string `json:"TLS_GREASE (0xfafa),omitempty"`
				X2551929        string `json:"X25519 (29),omitempty"`
			} `json:"shared_keys,omitempty"`
			PskKeyExchangeMode string   `json:"PSK_Key_Exchange_Mode,omitempty"`
			Versions           []string `json:"versions,omitempty"`
			Algorithms         []string `json:"algorithms,omitempty"`
		} `json:"extensions"`
		Version          string `json:"version"`
		Ja3NoPadding     string `json:"ja3_no_padding"`
		Ja3NoPaddingHash string `json:"ja3_no_padding_hash"`
		Ja3              string `json:"ja3"`
		Ja3Hash          string `json:"ja3_hash"`
		ClientRandom     string `json:"client_random"`
		SessionID        string `json:"session_id"`
		UsedTLSVersion   string `json:"used_tls_version"`
	} `json:"tls"`
	HTTP2 struct {
		AkamaiFingerprint     string `json:"akamai_fingerprint"`
		AkamaiFingerprintHash string `json:"akamai_fingerprint_hash"`
		SentFrames            []struct {
			FrameType string   `json:"frame_type"`
			Length    int      `json:"length"`
			Settings  []string `json:"settings,omitempty"`
			Increment int      `json:"increment,omitempty"`
			StreamID  int      `json:"stream_id,omitempty"`
			Headers   []string `json:"headers,omitempty"`
			Flags     []string `json:"flags,omitempty"`
			Priority  struct {
				Weight    int `json:"weight"`
				DependsOn int `json:"depends_on"`
				Exclusive int `json:"exclusive"`
			} `json:"priority,omitempty"`
		} `json:"sent_frames"`
	} `json:"http2"`
}

var browserFingerprints = map[string]map[string]map[string]string{
	chrome: {
		tls.HelloChrome_105.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			ja3StringWithPadding:  "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3HashWithPadding:    "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloChrome_104.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			ja3StringWithPadding:  "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3HashWithPadding:    "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloChrome_103.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			ja3StringWithPadding:  "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3HashWithPadding:    "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
	},
	firefox: {
		tls.HelloFirefox_102.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28,29-23-24-25-256-257,0",
			ja3Hash:               "e669667efb41c36f714c309243f41ca7",
			ja3StringWithPadding:  "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3HashWithPadding:    "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
		tls.HelloFirefox_104.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28,29-23-24-25-256-257,0",
			ja3Hash:               "e669667efb41c36f714c309243f41ca7",
			ja3StringWithPadding:  "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3HashWithPadding:    "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
	},
	opera: {
		tls.HelloOpera_89.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			ja3StringWithPadding:  "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3HashWithPadding:    "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloOpera_90.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			ja3StringWithPadding:  "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3HashWithPadding:    "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
	},
	safari: {
		tls.HelloSafari_15_3.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49188-49187-49162-49161-49192-49191-49172-49171-157-156-61-60-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "e476c7998ab30c28fa06df2fc809bd39",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49188-49187-49162-49161-49192-49191-49172-49171-157-156-61-60-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "c59b5aeb69936c251f090be89e1c4ca5",
			akamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
		tls.HelloSafari_15_6_1.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
		tls.HelloSafari_16_0.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
	},
	safariIpadOs: {
		tls.HelloIPad_15_6.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
	},
	safariIos: {
		tls.HelloIOS_15_5.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
		tls.HelloIOS_15_6.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
		tls.HelloIOS_16_0.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			ja3StringWithPadding:  "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3HashWithPadding:    "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
	},
}

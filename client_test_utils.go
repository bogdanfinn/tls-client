package tls_client

import tls "github.com/Carcraftz/utls"

const (
	Chrome    = "Chrome"
	Firefox   = "Firefox"
	Opera     = "Opera"
	Safari    = "Safari"
	SafariIOS = "Safari_IOS"

	Ja3String             = "ja3String"
	Ja3Hash               = "ja3Hash"
	AkamaiFingerprint     = "akamaiFingerprint"
	AkamaiFingerprintHash = "akamaiFingerprintHash"
)

var BrowserFingerprints = map[string]map[tls.ClientHelloID]map[string]string{
	Chrome: {
		tls.HelloChrome_103: map[string]string{
			Ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			Ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			AkamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			AkamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
	},
	Firefox: {
		tls.HelloFirefox_102: map[string]string{
			Ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28,29-23-24-25-256-257,0",
			Ja3Hash:               "e669667efb41c36f714c309243f41ca7",
			AkamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			AkamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
	},
	Opera: {
		tls.HelloOpera_89: map[string]string{
			Ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513,29-23-24,0",
			Ja3Hash:               "e1d8b04eeb8ef3954ec4f49267a783ef",
			AkamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			AkamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
	},
	Safari: {
		tls.HelloSafari_15_3: map[string]string{
			Ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49188-49187-49162-49161-49192-49191-49172-49171-157-156-61-60-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			Ja3Hash:               "e476c7998ab30c28fa06df2fc809bd39",
			AkamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			AkamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
		tls.HelloSafari_15_5: map[string]string{
			Ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			Ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			AkamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			AkamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
	},
	SafariIOS: {
		tls.HelloIOS_15_5: map[string]string{
			Ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27,29-23-24-25,0",
			Ja3Hash:               "f7854d0dd7148a99b75af38a7932fdec",
			AkamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			AkamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
	},
}

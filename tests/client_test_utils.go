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
	ja3Hash               = "ja3Hash"
	akamaiFingerprint     = "akamaiFingerprint"
	akamaiFingerprintHash = "akamaiFingerprintHash"
)

var browserFingerprints = map[string]map[string]map[string]string{
	chrome: {
		tls.HelloChrome_108.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,51-16-17513-43-11-65281-35-27-45-18-5-13-23-10-0-21,29-23-24,0",
			ja3Hash:               "9f97ee88116515aaa6cb4fd1caac7e16",
			akamaiFingerprint:     "1:65536,2:0,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "46cedabdca2073198a42fa10ca4494d0",
		},
		tls.HelloChrome_107.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,5-0-35-16-18-10-23-65281-43-51-27-17513-45-13-11-21,29-23-24,0",
			ja3Hash:               "a1ba8edd0661b9d9495ec40d4d3692e1",
			akamaiFingerprint:     "1:65536,2:0,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "46cedabdca2073198a42fa10ca4494d0",
		},
		tls.HelloChrome_106.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,2:0,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "46cedabdca2073198a42fa10ca4494d0",
		},
		tls.HelloChrome_105.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloChrome_104.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloChrome_103.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
	},
	firefox: {
		tls.HelloFirefox_102.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
		tls.HelloFirefox_104.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
		tls.HelloFirefox_105.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
		tls.HelloFirefox_106.Str(): map[string]string{
			ja3String:             "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
			ja3Hash:               "579ccef312d18482fc42e2b822ca2430",
			akamaiFingerprint:     "1:65536,4:131072,5:16384|12517377|3:0:0:201,5:0:0:101,7:0:0:1,9:0:7:1,11:0:3:1,13:0:0:241|m,p,a,s",
			akamaiFingerprintHash: "fd4f649c50a64e33cc9e2407055bafbe",
		},
	},
	opera: {
		tls.HelloOpera_89.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloOpera_90.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
		tls.HelloOpera_91.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24,0",
			ja3Hash:               "cd08e31494f9531f560d64c695473da9",
			akamaiFingerprint:     "1:65536,3:1000,4:6291456,6:262144|15663105|0|m,a,s,p",
			akamaiFingerprintHash: "7ad845f20fc17cc8088a0d9312b17da1",
		},
	},
	safari: {
		tls.HelloSafari_15_6_1.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
		tls.HelloSafari_16_0.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:4194304,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "e7b6dfd2eca81022e22f49765591e8c3",
		},
	},
	safariIpadOs: {
		tls.HelloIPad_15_6.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
	},
	safariIos: {
		tls.HelloIOS_15_5.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
		tls.HelloIOS_15_6.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
		tls.HelloIOS_16_0.Str(): map[string]string{
			ja3String:             "771,4865-4866-4867-49196-49195-52393-49200-49199-52392-49162-49161-49172-49171-157-156-53-47-49160-49170-10,0-23-65281-10-11-16-5-13-18-51-45-43-27-21,29-23-24-25,0",
			ja3Hash:               "773906b0efdefa24a7f2b8eb6985bf37",
			akamaiFingerprint:     "4:2097152,3:100|10485760|0|m,s,p,a",
			akamaiFingerprintHash: "8fe3e4ae51fb38d5c5108eabbf2a123c",
		},
	},
}

package profiles

import (
	"math"

	"github.com/bogdanfinn/fhttp/http2"
	tls "github.com/bogdanfinn/utls"
)

var MMSIos2 = getMMSClientProfile2()

func getMMSClientProfile2() ClientProfile {
	clientHelloId := tls.ClientHelloID{
		Client:  "MMSIos",
		Version: "2",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					0x1301,
					0x1302,
					0x1303,
					0xc02b,
					0xc02f,
					0xc02c,
					0xc030,
					0xcca9,
					0xcca8,
					0xc009,
					0xc013,
					0xc00a,
					0xc014,
					0x009c,
					0x009d,
					0x002f,
					0x0035,
					0x000a,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.CurveID(0x001d),
						tls.CurveID(0x0017),
						tls.CurveID(0x0018),
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []uint8{
						tls.PointFormatUncompressed,
					}},
					&tls.SessionTicketExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						0x0403,
						0x0804,
						0x0401,
						0x0503,
						0x0805,
						0x0501,
						0x0806,
						0x0601,
						0x0201,
					}},
					&tls.KeyShareExtension{KeyShares: []tls.KeyShare{
						{Group: tls.X25519},
					}},
					&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
						tls.PskModeDHE,
					}},
					&tls.SupportedVersionsExtension{Versions: []uint16{
						tls.VersionTLS13,
						tls.VersionTLS12,
					}},
				},
			}, nil
		},
	}

	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      4096,
		http2.SettingEnablePush:           1,
		http2.SettingMaxConcurrentStreams: 100,
		http2.SettingInitialWindowSize:    2097152,
		http2.SettingMaxFrameSize:         16384,
		http2.SettingMaxHeaderListSize:    math.MaxUint32,
	}

	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
		http2.SettingMaxHeaderListSize,
	}
	pseudoHeaderOrder := []string{
		":method",
		":scheme",
		":path",
		":authority",
	}

	return NewClientProfile(clientHelloId, settings, settingsOrder, pseudoHeaderOrder, 15663105, nil, nil)
}

var MMSIos3 = getMMSClientProfile3()

func getMMSClientProfile3() ClientProfile {
	clientHelloId := tls.ClientHelloID{
		Client:  "MMSIos",
		Version: "3",
		Seed:    nil,
		SpecFactory: func() (tls.ClientHelloSpec, error) {
			return tls.ClientHelloSpec{
				CipherSuites: []uint16{
					tls.GREASE_PLACEHOLDER,
					0x1301,
					0x1302,
					0x1303,
					0xc02c,
					0xc02b,
					0xcca9,
					0xc030,
					0xc02f,
					0xcca8,
					0xc00a,
					0xc009,
					0xc014,
					0xc013,
				},
				CompressionMethods: []uint8{
					tls.CompressionNone,
				},
				Extensions: []tls.TLSExtension{
					&tls.UtlsGREASEExtension{},
					&tls.SNIExtension{},
					&tls.ExtendedMasterSecretExtension{},
					&tls.RenegotiationInfoExtension{Renegotiation: tls.RenegotiateOnceAsClient},
					&tls.SupportedCurvesExtension{Curves: []tls.CurveID{
						tls.GREASE_PLACEHOLDER,
						0x001d,
						0x0017,
						0x0018,
						0x0019,
					}},
					&tls.SupportedPointsExtension{SupportedPoints: []uint8{
						tls.PointFormatUncompressed,
					}},
					&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
					&tls.StatusRequestExtension{},
					&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{
						0x0403,
						0x0804,
						0x0401,
						0x0503,
						0x0203,
						0x0805,
						0x0501,
						0x0806,
						0x0601,
						0x0201,
					}},
					&tls.SCTExtension{},
					&tls.KeyShareExtension{[]tls.KeyShare{
						{Group: tls.CurveID(tls.GREASE_PLACEHOLDER), Data: []byte{0}},
						{Group: tls.X25519},
					}},
					&tls.PSKKeyExchangeModesExtension{Modes: []uint8{
						tls.PskModeDHE,
					}},
					&tls.SupportedVersionsExtension{Versions: []uint16{
						tls.GREASE_PLACEHOLDER,
						tls.VersionTLS13,
						tls.VersionTLS12,
					}},
					&tls.UtlsCompressCertExtension{[]tls.CertCompressionAlgo{
						tls.CertCompressionZlib,
					}},
					&tls.UtlsGREASEExtension{},
					&tls.UtlsPaddingExtension{GetPaddingLen: tls.BoringPaddingStyle},
				},
			}, nil
		},
	}

	settings := map[http2.SettingID]uint32{
		http2.SettingHeaderTableSize:      4096,
		http2.SettingEnablePush:           1,
		http2.SettingMaxConcurrentStreams: 100,
		http2.SettingInitialWindowSize:    2097152,
		http2.SettingMaxFrameSize:         16384,
		http2.SettingMaxHeaderListSize:    math.MaxUint32,
	}

	settingsOrder := []http2.SettingID{
		http2.SettingHeaderTableSize,
		http2.SettingEnablePush,
		http2.SettingMaxConcurrentStreams,
		http2.SettingInitialWindowSize,
		http2.SettingMaxFrameSize,
		http2.SettingMaxHeaderListSize,
	}
	pseudoHeaderOrder := []string{
		":method",
		":scheme",
		":path",
		":authority",
	}

	return NewClientProfile(clientHelloId, settings, settingsOrder, pseudoHeaderOrder, 15663105, nil, nil)
}
